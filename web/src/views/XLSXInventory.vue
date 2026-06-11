<template>
  <v-container fluid>
    <v-card>
      <v-tabs v-model="activeFile" background-color="primary">
        <v-tab
          v-for="(file, index) in inventoryFiles"
          :key="index"
          :href="'#file-' + index"
        >
          {{ file.name }}
        </v-tab>
      </v-tabs>

      <v-tabs-items v-model="activeFile">
        <v-tab-item
          v-for="(file, index) in inventoryFiles"
          :key="index"
          :id="'file-' + index"
        >
          <v-tabs v-model="activeSheet[index]" class="mt-4">
            <v-tab
              v-for="(sheet, sIndex) in file.sheets"
              :key="sIndex"
              :href="'#sheet-' + index + '-' + sIndex"
            >
              {{ sheet.name }}
              <span class="ml-2 grey--text text--lighten-1">
                ({{ sheet.rows.length }} {{ $t('rows') }})
              </span>
            </v-tab>
          </v-tabs>

          <v-tabs-items v-model="activeSheet[index]">
            <v-tab-item
              v-for="(sheet, sIndex) in file.sheets"
              :key="sIndex"
              :id="'#sheet-' + index + '-' + sIndex"
            >
              <v-card-text>
                <v-text-field
                  v-model="searchQuery[index][sIndex]"
                  :placeholder="$t('search')"
                  prepend-icon="mdi-search"
                  class="mb-4"
                ></v-text-field>

                <v-data-table
                  :headers="getTableHeaders(sheet.headers)"
                  :items="getFilteredRows(sheet.rows, searchQuery[index][sIndex])"
                  class="elevation-1"
                  item-key="index"
                  :footer-props="{ 'items-per-page-options': [10, 20, 50] }"
                >
                  <template v-slot:item.ansible_password="{ item }">
                    <v-icon>mdi-eye-off</v-icon>
                    <span class="ml-2">{{ maskPassword(item.ansible_password) }}</span>
                  </template>
                </v-data-table>

                <v-chip
                  v-if="isHostsSheet(sheet)"
                  color="blue"
                  class="mt-4"
                >
                  {{ $t('hostInventory') }}
                </v-chip>
                <v-chip
                  v-else-if="isGroupVarsSheet(sheet)"
                  color="green"
                  class="mt-4"
                >
                  {{ $t('groupVariables') }}
                </v-chip>
              </v-card-text>
            </v-tab-item>
          </v-tabs-items>
        </v-tab-item>
      </v-tabs-items>
    </v-card>
  </v-container>
</template>

<script>
import axios from 'axios';

export default {
  name: 'XLSXInventory',
  data() {
    return {
      inventoryFiles: [],
      activeFile: 0,
      activeSheet: [],
      searchQuery: [],
    };
  },
  mounted() {
    this.loadInventory();
  },
  methods: {
    async loadInventory() {
      try {
        const response = await axios.get(`/api/project/${this.$route.params.projectId}/xlsx-inventory`);
        this.inventoryFiles = response.data;
        this.activeSheet = this.inventoryFiles.map(() => 0);
        this.searchQuery = this.inventoryFiles.map((file) => file.sheets.map(() => ''));
      } catch (error) {
        console.error('Failed to load inventory:', error);
      }
    },
    getTableHeaders(headers) {
      return headers.map((h) => ({
        text: h,
        value: h,
      }));
    },
    getFilteredRows(rows, query) {
      if (!query) return rows;
      const q = query.toLowerCase();
      return rows.filter((row) => (
        Object.values(row).some((val) => String(val).toLowerCase().includes(q))
      ));
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
  },
};
</script>
