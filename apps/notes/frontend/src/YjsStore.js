import Store from 'orbit-db-store';
import { mergeUpdates } from 'yjs';
class YIndex {
  get() {
    return this._update
  }

  updateIndex(oplog) {
    console.log()
    const updates = oplog.values.map(op => op.payload)
    console.log(updates)
    const merged =  mergeUpdates(updates)
    console.log(merged)
    this._update = merged
  }
}

export default class KeyValueStore extends Store {
  constructor(ipfs, identity, address, options) {
    Object.assign(options || {}, { Index: YIndex });
    super(ipfs, identity, address, options)
  }

  get() {
    return this._index.get();
  }

  set(key, data) {
    this.put(key, data);
  }

  put(update) {
    return this._addOperation(update);
  }
}