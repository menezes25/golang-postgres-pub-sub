module.exports = {
  // lintOnSave: false,
  publicPath: "",
  chainWebpack: config => {
    // vue inspect --plugin html
    // Alterar titulo do html
    config.plugin('html').tap((args) => {
      args[0].title = 'POC - Postgres PUB/SUB';
      return args;
    });
  },
};
