const fs = require("fs");
const data = fs.readFileSync("/etc/hostname", "utf8");
console.log(data);
