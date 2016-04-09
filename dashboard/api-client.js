'use strict';

function cachedApiJson({ cache, cacheKey, endpoint }) {
  let promiseForResult = cache.get(cacheKey);
  if (promiseForResult) {
    return promiseForResult;
  }
  promiseForResult = fetch(endpoint)
    .then((response) => { return response.json(); })
    .catch((err) => {
      cache.delete(cacheKey);
      return Promise.reject(err);
    });
  cache.set(cacheKey, promiseForResult);
  return promiseForResult;
}

export default class ApiClient {
  constructor({ cache }) {
    this._cache = cache;
  }

  listStarCounts({ owner, repo }) {
    return cachedApiJson({
      cache: this._cache,
      cacheKey: `gh:${owner}:${repo}:star_counts`,
      endpoint: `/gh/${owner}/${repo}/star_counts`,
    });
  }

  listIssueCounts({ owner, repo }) {
    return cachedApiJson({
      cache: this._cache,
      cacheKey: `gh:${owner}:${repo}:issue_counts`,
      endpoint: `/gh/${owner}/${repo}/issue_counts`,
    });
  }

  listTopIssues({ owner, repo }) {
    return cachedApiJson({
      cache: this._cache,
      cacheKey: `gh:${owner}:${repo}:top_issues`,
      endpoint: `/gh/${owner}/${repo}/top_issues`,
    });
  }

  listTopPrs({ owner, repo }) {
    return cachedApiJson({
      cache: this._cache,
      cacheKey: `gh:${owner}:${repo}:top_prs`,
      endpoint: `/gh/${owner}/${repo}/top_prs`,
    });
  }

  listTopContributors({ owner, repo, date }) {
    const year = ''+date.getFullYear();
    const month = ('0' + (date.getMonth()+1)).slice(-2);

    return cachedApiJson({
      cache: this._cache,
      cacheKey: `gh:${owner}:${repo}:highscores:${year}:${month}`,
      endpoint: `/gh/${owner}/${repo}/highscores/${year}/${month}`,
    });
  }
}
