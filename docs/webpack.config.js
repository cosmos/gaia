const path = require('path');

module.exports = {
  mode: 'development',
  optimization: {
    minimize: false,
  },
  output: {
    clean: true, // Clean the output directory before emit.
  },  
};
