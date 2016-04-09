'use strict';

var flatten = require('lodash.flatten'),
  webpack = require('webpack');

/*global __dirname, process*/

exports.entry = './index.js';

exports.output = {
  filename: 'bundle.min.js',
  path: __dirname,
};

exports.module = {
  loaders: [{
    exclude: /node_modules/,
    loader: 'babel',
    query: { presets: ['es2015', 'react', 'stage-2'] },
    test: /\.js$/,
  }],
};

exports.plugins = flatten([
  [new webpack.EnvironmentPlugin(['NODE_ENV'])],

  (process.env.NODE_ENV === 'production' ?
     [new webpack.optimize.UglifyJsPlugin({ compress: { warnings: false } })] :
     []),
]);

exports.devtool = (process.env.NODE_ENV === 'production' ?
                     'source-map' :
                     'eval-source-map');
