import Cache from '../cache';
import { expect } from 'chai';
import constant from 'lodash.constant';

describe('Cache', () => {
  let cache;

  beforeEach(() => {
    cache = new Cache({ maxAge: 1000 * 60 * 5 });
  });

  it('gets undefined for a non-existing key', () => {
    expect(cache.get('nonexisting-key')).to.be.undefined;
  });

  it('sets and gets an item', () => {
    cache.set('fish-key', { name: 'fish-value' });
    expect(cache.get('fish-key')).to.eql({ name: 'fish-value' });
  });

  it('expires objects older than the maxAge', () => {
    const oldDateNow = Date.now;
    Date.now = constant(0);
    cache.set('expiring-key', 'some-nonsense');
    Date.now = constant(1000 * 60 * 5 + 1);
    expect(cache.get('expiring-key')).to.be.undefined;
    Date.now = oldDateNow;
  });

  it('deletes an item', () => {
    cache.set('deleted-key', 'some-nonsense');
    cache.delete('deleted-key');
    expect(cache.get('deleted-key')).to.be.undefined;
  });
});
