const fs = require("fs");
fs.readFile("/etc/hostname", "utf8", (err, data) => {
  console.log(data);
});
