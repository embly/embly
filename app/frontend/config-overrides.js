module.exports = {
  webpack: function(config, env) {
    config.output.publicPath = "/app";
    return config;
  }
};
