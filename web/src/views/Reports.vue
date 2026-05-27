<!-- web/src/views/Reports.vue -->
<template>
  <div>
    <v-card>
      <!-- 面包屑导航 -->
      <v-card-title>
        <v-breadcrumbs class="mb-2">
          <v-breadcrumb-item
            v-for="(item, index) in breadcrumbs"
            :key="index"
            @click="navigateToPath(item.path)"
            :disabled="index === breadcrumbs.length - 1"
          >
            {{ item.name }}
          </v-breadcrumb-item>
        </v-breadcrumbs>
        {{ $t('reports') }}
      </v-card-title>

      <v-card-text>
        <v-data-table
          :headers="headers"
          :items="files"
          :loading="loading"
          class="elevation-1"
        >
          <template v-slot:item.name="{ item }">
            <v-list-item
              @click="handleItemClick(item)"
              class="cursor-pointer"
            >
              <v-icon v-if="item.is_dir">mdi-folder</v-icon>
              <v-icon v-else>mdi-file</v-icon>
              <span class="ml-2">{{ item.name }}</span>
            </v-list-item>
          </template>
          <template v-slot:item.size="{ item }">
            {{ formatSize(item.size) }}
          </template>
          <template v-slot:item.actions="{ item }">
            <v-btn
              v-if="!item.is_dir"
              icon
              color="primary"
              @click="downloadFile(item.name)"
            >
              <v-icon>mdi-download</v-icon>
            </v-btn>
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
      currentPath: '',
      breadcrumbs: [],
      headers: [
        { text: this.$t('filename'), value: 'name' },
        { text: this.$t('size'), value: 'size' },
        { text: this.$t('modified'), value: 'mod_time' },
        { text: this.$t('actions'), value: 'actions', sortable: false },
      ],
    };
  },
  mounted() {
    this.loadFiles();
  },
  methods: {
    async loadFiles() {
      this.loading = true;
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
        // 进入子目录
        const newPath = this.currentPath ? `${this.currentPath}/${item.name}` : item.name;
        this.currentPath = newPath;
        this.loadFiles();
      }
    },
    navigateToPath(path) {
      this.currentPath = path;
      this.loadFiles();
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
    downloadFile(filename) {
      const fullPath = this.currentPath ? `${this.currentPath}/${filename}` : filename;
      const link = document.createElement('a');
      link.href = `/api/project/${this.$route.params.projectId}/reports/download/${encodeURIComponent(fullPath)}`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
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
