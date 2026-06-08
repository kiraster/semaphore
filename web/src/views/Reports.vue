<!-- web/src/views/Reports.vue -->
<template>
  <div>
    <v-card>
      <v-card-title class="align-center">
        <v-btn
          v-if="currentPath"
          icon
          @click="goBack"
          class="mr-2"
          depressed
        >
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>

        <v-breadcrumbs class="flex-grow-1" :items="breadcrumbs">
          <template v-slot:item="{ item }">
            <v-breadcrumb-item
              @click="navigateToPath(item.path)"
              :disabled="item.path === currentPath"
            >
              {{ item.name }}
            </v-breadcrumb-item>
          </template>
        </v-breadcrumbs>

        <v-btn
          v-if="selectedFiles.length > 0"
          color="primary"
          @click="downloadSelectedAsZip"
          :loading="isDownloading"
        >
          <v-icon left>mdi-download</v-icon>
          批量下载 ({{ selectedFiles.length }})
        </v-btn>
      </v-card-title>

      <v-card-text>
        <v-data-table
          :headers="headers"
          :items="files"
          :loading="loading"
          class="elevation-1 report-table"
          item-key="name"
          :footer-props="{ 'items-per-page-options': [10, 20, 50, 100] }"
        >
          <template v-slot:item.checkbox="{ item }">
            <v-checkbox
              v-if="!item.is_dir"
              v-model="selectedFiles"
              :value="item.name"
              hide-details
              class="checkbox-cell"
            ></v-checkbox>
          </template>

          <template v-slot:item.name="{ item }">
            <v-list-item @click="handleItemClick(item)" class="cursor-pointer">
              <v-icon v-if="item.is_dir">mdi-folder</v-icon>
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
                v-if="!item.is_dir"
                icon
                color="primary"
                @click="downloadSingleFile(item.name)"
              >
                <v-icon>mdi-download</v-icon>
              </v-btn>
            </div>
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
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
      breadcrumbs: [],
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
  methods: {
    async loadFiles() {
      this.loading = true;
      this.selectedFiles = [];
      try {
        const pathParam = this.currentPath ? `?path=${encodeURIComponent(this.currentPath)}` : '';
        const response = await axios.get(`/api/project/${this.$route.params.projectId}/reports${pathParam}`);
        this.files = response.data;
        this.updateBreadcrumbs();
      } catch (error) {
        console.error('Failed to load report files:', error);
      } finally {
        this.loading = false;
      }
    },
    handleItemClick(item) {
      if (item.is_dir) {
        const newPath = this.currentPath ? `${this.currentPath}/${item.name}` : item.name;
        this.currentPath = newPath;
        this.loadFiles();
      }
    },
    navigateToPath(path) {
      this.currentPath = path;
      this.loadFiles();
    },
    goBack() {
      if (this.currentPath) {
        const parts = this.currentPath.split('/');
        this.currentPath = parts.slice(0, -1).join('/');
        this.loadFiles();
      }
    },
    updateBreadcrumbs() {
      this.breadcrumbs = [{ name: '报告', path: '' }];
      if (this.currentPath) {
        const parts = this.currentPath.split('/');
        let currentPath = '';
        parts.forEach((part) => {
          currentPath = currentPath ? `${currentPath}/${part}` : part;
          this.breadcrumbs.push({ name: part, path: currentPath });
        });
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
        console.error('Failed to download zip:', error);
        this.$snackbar({
          color: 'error',
          text: '批量下载失败，请重试',
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
