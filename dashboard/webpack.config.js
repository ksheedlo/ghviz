'use strict';

var webpack = require('webpack');

/*global __dirname*/
module.exports = {
  entry: './index.js',
  output: {
    filename: 'bundle.min.js',
    path: __dirname
  },
  module: {
    loaders: [{
      exclude: /node_modules/,
      loader: 'babel',
      query: { presets: ['es2015', 'react', 'stage-2'] },
      test: /\.js$/
    }]
  },
  plugins: [
    new webpack.EnvironmentPlugin(['NODE_ENV']),
    new webpack.optimize.UglifyJsPlugin({ compress: { warnings: false } })
  ]
};
