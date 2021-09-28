const p = require('../current-parameters.json');

module.exports = {
    theme: "cosmos",
    title: "Cosmos Hub Governance",
    base: '/governance/',
    patterns: ['**/*.md', '**/*.vue', '!scripts/**'],
    themeConfig:{
        custom: true,
        topbar: {
            banner: false,
        },
        currentParameters: p,
        sidebar: {
            auto: true
        }
    }
 };