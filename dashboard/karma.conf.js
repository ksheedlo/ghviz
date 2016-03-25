var webpackConfig = require('./webpack.config.js');
webpackConfig.entry = {};

module.exports = function(config) {
  config.set({
    autoWatch: true,
    basePath: '',
    browsers: ['Chrome', 'Firefox'],
    colors: true,
    concurrency: Infinity,
    exclude: [],
    frameworks: ['mocha'],
    files: [
      'third_party/sinon/1.17.3/sinon-1.17.3.js',

      'test/**/*.js'
    ],
    logLevel: config.LOG_INFO,
    port: 9876,
    preprocessors: {
      'test/**/*.js': ['webpack']
    },
    reporters: ['progress'],
    singleRun: false,
    webpack: webpackConfig,
    webpackMiddleware: { noInfo: true }
  });
};
