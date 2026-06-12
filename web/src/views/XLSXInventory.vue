<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <!-- 加载状态 -->
  <div v-if="!isLoaded">
    <v-progress-linear indeterminate color="primary darken-2"></v-progress-linear>
  </div>
  <!-- 主内容 -->
  <div v-else>
    <!-- 工具栏 -->
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('xlsxInventory') }}</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>
    <!-- 文件选择标签页 -->
    <v-tabs show-arrows class="pl-4" v-model="activeFile">
      <v-tab
        v-for="(file, index) in inventoryFiles"
        :key="index"
      >
        {{ file.name }}
      </v-tab>
    </v-tabs>
    <v-divider style="margin-top: -1px;"/>
    <!-- Sheet 选择标签页 -->
    <v-tabs show-arrows class="pl-4" v-model="activeSheet">
      <v-tab
        v-for="(sheet, index) in currentFile?.sheets || []"
        :key="index"
      >
        {{ sheet.name }} ({{ sheet.rows.length }} 行)
      </v-tab>
    </v-tabs>
    <v-divider style="margin-top: -1px;"/>

    <!-- 搜索框 -->
    <div class="pl-4 pr-4 mt-4">
      <v-text-field
        v-model="searchQuery"
        placeholder="搜索..."
        prepend-icon="mdi-search"
      ></v-text-field>
    </div>
    <!-- 数据表格 -->
    <v-data-table
      class="mt-4"
      :headers="tableHeaders"
      :items="tableItems"
      item-key="__id"
      :footer-props="{ 'items-per-page-options': [10, 20, 50] }"
    >
      <template v-slot:item.ansible_password="{ item }">
        <v-icon>mdi-eye-off</v-icon>
        <span class="ml-2">{{ maskPassword(item.ansible_password) }}</span>
      </template>
      <template v-slot:item.password="{ item }">
        <v-icon>mdi-eye-off</v-icon>
        <span class="ml-2">{{ maskPassword(item.password) }}</span>
      </template>
    </v-data-table>
    <!-- Sheet 类型标识 -->
    <div class="pl-4 pb-4">
      <v-chip v-if="isHostsSheet(currentSheet)" color="blue">
        主机清单
      </v-chip>
      <v-chip v-else-if="isGroupVarsSheet(currentSheet)" color="green">
        组变量
      </v-chip>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'XLSXInventory',
  data() {
    return {
      inventoryFiles: [],
      activeFile: 0,
      activeSheet: 0,
      searchQuery: '',
      isLoaded: false,
    };
  },
  mounted() {
    console.log('XLSXInventory mounted');
    this.loadInventory();
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    currentFile() {
      return this.inventoryFiles[this.activeFile] || null;
    },
    currentSheet() {
      if (!this.currentFile) return null;
      return this.currentFile.sheets[this.activeSheet] || null;
    },
    tableHeaders() {
      if (!this.currentSheet) return [];
      return this.currentSheet.headers.map((h) => ({ text: h, value: h }));
    },
    tableItems() {
      if (!this.currentSheet) return [];
      let items = this.currentSheet.rows.map((row, index) => ({ ...row, __id: index }));
      if (this.searchQuery) {
        const q = this.searchQuery.toLowerCase();
        items = items.filter((row) => (
          Object.values(row).some((val) => String(val).toLowerCase().includes(q))
        ));
      }
      return items;
    },
  },
  methods: {
    async loadInventory() {
      try {
        console.log('Loading inventory for project:', this.projectId);
        const response = await axios.get(`/api/project/${this.projectId}/xlsx-inventory`);
        this.inventoryFiles = response.data;
        console.log('Inventory loaded:', this.inventoryFiles);
      } catch (error) {
        console.error('Failed to load inventory:', error);
      } finally {
        this.isLoaded = true;
        console.log('isLoaded set to true');
      }
    },
    maskPassword(password) {
      if (!password) return '';
      if (password.length <= 4) return '****';
      return password[0] + '*'.repeat(password.length - 2) + password.slice(-1);
    },
    isHostsSheet(sheet) {
      const hostHeaders = ['hostname', 'host', 'ansible_host', 'name'];
      return sheet.headers.some((h) => hostHeaders.includes(h.toLowerCase()));
    },
    isGroupVarsSheet(sheet) {
      const groupHeaders = ['group', 'group_name', 'ansible_connection'];
      return sheet.headers.some((h) => groupHeaders.includes(h.toLowerCase()));
    },
    showDrawer() {
      if (this.$store) {
        this.$store.commit('toggleDrawer');
      }
    },
  },
};
</script>
