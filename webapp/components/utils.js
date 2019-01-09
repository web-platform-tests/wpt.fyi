function ensureTrailingSlash(path) {
  return path.endsWith('/') ? path : (path + '/');
}
export { ensureTrailingSlash };