package internal

// VERSION 版本号。用 var 而非 const，以便构建时通过
// -ldflags "-X minijvm/internal.VERSION=x.y.z" 注入。
var VERSION = "0.0.1"
