'use strict';

const Cache = require('./cache');

const cache = new Cache({ maxAge: 1000 * 60 * 5 });

exports.listStarCounts = function listStarCounts({ owner, repo }) {
  const cacheKey = `gh:${owner}:${repo}:star_counts`;

  let promiseForStarCounts = cache.get(cacheKey);
  if (promiseForStarCounts) {
    return promiseForStarCounts;
  }
  promiseForStarCounts = fetch(`/gh/${owner}/${repo}/star_counts`)
    .then((response) => { return response.json(); });
  cache.set(cacheKey, promiseForStarCounts);
  return promiseForStarCounts;
};

exports.listIssueCounts = function listStarCounts({ owner, repo }) {
  const cacheKey = `gh:${owner}:${repo}:issue_counts`;

  let promiseForIssueCounts = cache.get(cacheKey);
  if (promiseForIssueCounts) {
    return promiseForIssueCounts;
  }
  promiseForIssueCounts = fetch(`/gh/${owner}/${repo}/issue_counts`)
    .then((response) => { return response.json(); });
  cache.set(cacheKey, promiseForIssueCounts);
  return promiseForIssueCounts;
};

exports.listTopIssues = function listTopIssues({ owner, repo }) {
  const cacheKey = `gh:${owner}:${repo}:top_issues`;

  let promiseForTopIssues = cache.get(cacheKey);
  if (promiseForTopIssues) {
    return promiseForTopIssues;
  }
  promiseForTopIssues = fetch(`/gh/${owner}/${repo}/top_issues`)
    .then((response) => { return response.json(); });
  cache.set(cacheKey, promiseForTopIssues);
  return promiseForTopIssues;
};

exports.listTopPrs = function listTopPrs({ owner, repo }) {
  const cacheKey = `gh:${owner}:${repo}:top_prs`;

  let promiseForTopPrs = cache.get(cacheKey);
  if (promiseForTopPrs) {
    return promiseForTopPrs;
  }
  promiseForTopPrs = fetch(`/gh/${owner}/${repo}/top_prs`)
    .then((response) => { return response.json(); });
  cache.set(cacheKey, promiseForTopPrs);
  return promiseForTopPrs;
};

exports.listTopContributors = function listTopContributors({ owner, repo, date }) {
  const year = ''+date.getFullYear();
  const month = ('0' + (date.getMonth()+1)).slice(-2);
  const cacheKey = `gh:${owner}:${repo}:highscores:${year}:${month}`;

  let promiseForTopContributors = cache.get(cacheKey);
  if (promiseForTopContributors) {
    return promiseForTopContributors;
  }
  promiseForTopContributors = fetch(`/gh/${owner}/${repo}/highscores/${year}/${month}`)
    .then((response) => { return response.json(); });
  cache.set(cacheKey, promiseForTopContributors);
  return promiseForTopContributors;
};
