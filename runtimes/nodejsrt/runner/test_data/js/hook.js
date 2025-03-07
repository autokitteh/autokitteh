const fs = require('fs');

function listDir() {
    return fs.readdirSync(".")
}

results = listDir()
