'use strict';

var path = require('path');
var fs = require('fs');
var _ = require('lodash');

/*global __dirname*/

exports.externals = _.fromPairs(_.map(
  _.filter(fs.readdirSync('node_modules'), function (dir) {
    return ['.bin'].indexOf(dir) === -1;
  }),
  function (module) {
    return [module, 'commonjs ' + module];
  })
);

exports.entry = './server.js';

exports.target = 'node';

exports.output = {
  path: path.join(__dirname, 'dist'),
  filename: 'server.js',
};

exports.module = {
  loaders: [{
    exclude: /node_modules/,
    loader: 'babel',
    query: { presets: ['es2015', 'react', 'stage-2'] },
    test: /\.js$/,
  }],
};
