<!-- web/src/views/Reports.vue -->
<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <!-- 加载状态 -->
  <div v-if="loading">
    <v-progress-linear
      indeterminate
      color="primary darken-2"
    ></v-progress-linear>
  </div>

  <div v-else>
    <!-- 工具栏（包含面包屑） -->
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>

      <!-- 页面标题 + 路径显示 -->
      <div class="flex-grow-1 d-flex align-center">
        <span class="text-h6 font-weight-bold text-primary">{{ $t('reports') }}</span>
        <span
         v-if="currentPath" class="text-h6 font-weight-bold text-primary">/{{ currentPath }}/
        </span>
      </div>

      <!-- 批量下载按钮 -->
      <v-btn
        v-if="selectedFiles.length > 0"
        color="primary"
        @click="downloadSelectedAsZip"
        :loading="isDownloading"
      >
        <v-icon left>mdi-download</v-icon>
        {{ $t('download') }} ({{ selectedFiles.length }})
      </v-btn>
    </v-toolbar>

    <!-- 分隔线 -->
    <v-divider />

    <!-- 数据表格 -->
    <v-data-table
      :headers="headers"
      :items="displayFiles"
      class="mt-4 elevation-1 report-table"
      item-key="__id"
      :footer-props="{ 'items-per-page-options': [10, 20, 50, 100] }"
    >
      <template v-slot:item.checkbox="{ item }">
        <v-checkbox
          v-if="!item.is_parent"
          v-model="selectedFiles"
          :value="item.name"
          hide-details
          class="checkbox-cell"
        ></v-checkbox>
      </template>

      <template v-slot:item.name="{ item }">
        <v-list-item @click="handleItemClick(item)" class="cursor-pointer">
          <v-icon v-if="item.is_parent">mdi-folder-up</v-icon>
          <v-icon v-else-if="item.is_dir">mdi-folder</v-icon>
          <v-icon v-else>mdi-file</v-icon>
          <span class="ml-2">{{ item.name }}</span>
        </v-list-item>
      </template>

      <template v-slot:item.size="{ item }">
        <span class="text-center d-block">{{ formatSize(item.size) }}</span>
      </template>

      <template v-slot:item.mod_time="{ item }">
        <span class="text-center d-block">{{ item.mod_time }}</span>
      </template>

      <template v-slot:item.actions="{ item }">
        <div class="text-center">
          <v-btn
            v-if="!item.is_parent"
            icon
            color="primary"
            @click="downloadItem(item)"
          >
            <v-icon>mdi-download</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Reports',
  data() {
    return {
      files: [],
      loading: false,
      isDownloading: false,
      currentPath: '',
      selectedFiles: [],
      headers: [
        {
          text: '',
          value: 'checkbox',
          width: '40px',
          sortable: false,
          align: 'center',
        },
        { text: this.$t('filename'), value: 'name' },
        {
          text: this.$t('size'),
          value: 'size',
          width: '100px',
          sortable: true,
          align: 'center',
        },
        {
          text: this.$t('modified'),
          value: 'mod_time',
          width: '160px',
          sortable: true,
          align: 'center',
        },
        {
          text: this.$t('actions'),
          value: 'actions',
          width: '80px',
          sortable: false,
          align: 'center',
        },
      ],
    };
  },
  mounted() {
    this.loadFiles();
  },
  computed: {
    displayFiles() {
      if (!this.files) {
        return [];
      }

      const filesWithId = this.files.map((file, index) => ({
        ...file,
        __id: `${file.name}_${index}`,
      }));

      if (this.currentPath) {
        filesWithId.unshift({
          name: 'Parent directory/',
          size: 0,
          mod_time: '',
          is_dir: false,
          is_parent: true,
          __id: 'parent_directory',
        });
      }

      return filesWithId;
    },
  },
  methods: {
    showDrawer() {
      if (this.$store) {
        this.$store.commit('toggleDrawer');
      }
    },
    async loadFiles() {
      this.loading = true;
      this.selectedFiles = [];

      try {
        const projectId = this.$route?.params?.projectId;
        if (!projectId) {
          throw new Error('Project ID not found');
        }

        const pathParam = this.currentPath ? `?path=${encodeURIComponent(this.currentPath)}` : '';
        const response = await axios.get(`/api/project/${projectId}/reports${pathParam}`);
        this.files = response.data || [];
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Failed to load report files:', error);
        this.files = [];
      } finally {
        this.loading = false;
      }
    },
    handleItemClick(item) {
      if (item.is_parent) {
        this.goBack();
      } else if (item.is_dir) {
        const newPath = this.currentPath ? `${this.currentPath}/${item.name}` : item.name;
        this.currentPath = newPath;
        this.loadFiles();
      }
    },
    goBack() {
      if (this.currentPath) {
        const parts = this.currentPath.split('/');
        this.currentPath = parts.slice(0, -1).join('/');
        this.loadFiles();
      }
    },
    downloadItem(item) {
      if (item.is_dir) {
        this.downloadDirectoryAsZip(item.name);
      } else {
        this.downloadSingleFile(item.name);
      }
    },
    downloadSingleFile(filename) {
      const fullPath = this.currentPath ? `${this.currentPath}/${filename}` : filename;
      const link = document.createElement('a');
      link.href = `/api/project/${this.$route.params.projectId}/reports/download/${encodeURIComponent(fullPath)}`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
    async downloadDirectoryAsZip(dirname) {
      this.isDownloading = true;

      try {
        const dirPath = this.currentPath ? `${this.currentPath}/${dirname}` : dirname;
        const response = await axios.post(
          `/api/project/${this.$route.params.projectId}/reports/download/zip`,
          { files: [dirPath] },
          { responseType: 'blob' },
        );

        const url = window.URL.createObjectURL(response.data);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${dirname}.zip`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Failed to download directory:', error);
        this.$snackbar({
          color: 'error',
          text: this.$t('downloadFailed'),
        });
      } finally {
        this.isDownloading = false;
      }
    },
    async downloadSelectedAsZip() {
      if (this.selectedFiles.length === 0) return;

      this.isDownloading = true;

      try {
        const filesWithPath = this.selectedFiles.map((filename) => (
          this.currentPath ? `${this.currentPath}/${filename}` : filename
        ));

        const response = await axios.post(
          `/api/project/${this.$route.params.projectId}/reports/download/zip`,
          { files: filesWithPath },
          { responseType: 'blob' },
        );

        const url = window.URL.createObjectURL(response.data);
        const link = document.createElement('a');
        link.href = url;
        link.download = `reports_${Date.now()}.zip`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);

        this.selectedFiles = [];
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Failed to download zip:', error);
        this.$snackbar({
          color: 'error',
          text: this.$t('downloadFailed'),
        });
      } finally {
        this.isDownloading = false;
      }
    },
    formatSize(bytes) {
      if (bytes === 0) {
        return '0 B';
      }
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return `${parseFloat((bytes / (k ** i)).toFixed(2))} ${sizes[i]}`;
    },
  },
};
</script>

<style scoped>
.report-table .checkbox-cell {
  padding: 0;
  margin: 0;
}

.report-table .v-data-table__td {
  padding: 8px 12px;
}

.report-table .v-data-table__th {
  padding: 8px 12px;
  text-align: center;
}
</style>
