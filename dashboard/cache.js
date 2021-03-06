export default class Cache {
  constructor({ maxAge }) {
    this._maxAge = maxAge;
    this._cache = {};
  }

  get(key) {
    if (!this._cache[key]) {
      return (void 0);
    }
    if (this._cache[key].expires < Date.now()) {
      delete this._cache[key];
      return (void 0);
    }
    return this._cache[key].value;
  }

  set(key, value) {
    this._cache[key] = { value,
                         expires: Date.now() + this._maxAge };
  }

  delete(key) {
    delete this._cache[key];
  }
}
