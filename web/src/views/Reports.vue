<!-- web/src/views/Reports.vue -->
<template>
  <div>
    <v-card>
      <!-- 操作栏 -->
      <v-card-title>
        <!-- 返回按钮 -->
        <v-btn
          v-if="currentPath"
          icon
          @click="goBack"
          class="mr-4"
        >
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        
        <!-- 面包屑导航 -->
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
        
        <!-- 批量下载按钮 -->
        <v-spacer />
        <v-btn
          v-if="selectedFiles.length > 0"
          color="primary"
          @click="downloadSelected"
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
          class="elevation-1"
          item-key="name"
        >
          <!-- 复选框列 -->
          <template v-slot:item.checkbox="{ item }">
            <v-checkbox
              v-if="!item.is_dir"
              v-model="selectedFiles"
              :value="item.name"
              hide-details
            ></v-checkbox>
          </template>
          
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
      selectedFiles: [],
      headers: [
        { text: '', value: 'checkbox', sortable: false },
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
      this.selectedFiles = []; // 重置选择
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
      // 返回上一级目录
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
    downloadFile(filename) {
      const fullPath = this.currentPath ? `${this.currentPath}/${filename}` : filename;
      const link = document.createElement('a');
      link.href = `/api/project/${this.$route.params.projectId}/reports/download/${encodeURIComponent(fullPath)}`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
    downloadSelected() {
      // 逐个下载选中的文件
      this.selectedFiles.forEach((filename) => {
        setTimeout(() => {
          this.downloadFile(filename);
        }, 500); // 延迟下载，避免浏览器阻止
      });
      this.selectedFiles = []; // 清空选择
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
