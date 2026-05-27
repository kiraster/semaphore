<!-- web/src/views/Reports.vue -->
<template>
  <div>
    <v-card>
      <v-card-title>{{ $t('reports') }}</v-card-title>
      <v-card-text>
        <v-data-table
          :headers="headers"
          :items="files"
          :loading="loading"
          class="elevation-1"
        >
          <template v-slot:item.name="{ item }">
            <v-icon v-if="item.is_dir">mdi-folder</v-icon>
            <v-icon v-else>mdi-file</v-icon>
            <span class="ml-2">{{ item.name }}</span>
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
        const response = await axios.get(`/api/project/${this.$route.params.projectId}/reports`);
        this.files = response.data;
      } catch (error) {
        console.error('Failed to load report files:', error);
      } finally {
        this.loading = false;
      }
    },
    downloadFile(filename) {
      // 创建下载链接
      const link = document.createElement('a');
      link.href = `/api/project/${this.$route.params.projectId}/reports/${encodeURIComponent(filename)}`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
    formatSize(bytes) {
      if (bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return `${parseFloat((bytes / (k ** i)).toFixed(2))} ${sizes[i]}`;
    },
  },
};
</script>
