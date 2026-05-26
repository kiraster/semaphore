const webpack = require('webpack');

module.exports = {
  configureWebpack: {
    plugins: [
      new webpack.DefinePlugin({
        'process.env.VUE_APP_BUILD_TYPE': JSON.stringify(process.env.VUE_APP_BUILD_TYPE),
      }),
    ],
    devServer: {
      historyApiFallback: true,
      proxy: {
        '^/api': {
          // 调试前端页面指定后端api服务器2026-05-25 14:13:20
          // target: 'http://localhost:3000',
          target: 'http://172.31.100.251:3000',
        },
      },
    },
  },
  chainWebpack: (config) => {
    config.plugin('html')
      .tap((args) => {
        // eslint-disable-next-line no-param-reassign
        args[0].minify = false;
        return args;
      });
  },
  transpileDependencies: [
    'vuetify',
  ],
  publicPath: './',
  outputDir: '../api/public',
};
