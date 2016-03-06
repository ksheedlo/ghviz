'use strict';

const bluebird = require('bluebird');
if (!window.Promise) {
  window.Promise = bluebird;
}
require('whatwg-fetch');

const STAR_COUNT_SELECTOR = '.tile__star-count';
const ISSUE_COUNT_SELECTOR = '.tile__issue-count';
const PR_COUNT_SELECTOR = '.tile__pr-count';

const LINE_CHART_MARGIN = {
  top: 20,
  right: 20,
  bottom: 30,
  left: 50
};

const d3 = require('d3'),
  map = require('lodash.map'),
  $ = require('jquery');

fetch('/gh/rackerlabs/mimic/star_counts').then((resp) => {
  return resp.json();
})
.then((starCounts) => {
  const formattedCounts = map(starCounts, (starCount) => {
    return { stars: starCount.Stars,
             timestamp: d3.time.format.iso.parse(starCount.Timestamp) };
  });

  const starCountEl = $(STAR_COUNT_SELECTOR);

  const height =
    starCountEl.height() - (LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom);
  const width =
    starCountEl.width() - (LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right);

  const t = d3.time.scale()
    .range([0, width]);

  const tAxis = d3.svg.axis()
    .scale(t)
    .orient('bottom');

  const y = d3.scale.linear()
    .range([height, 0]);

  const yAxis = d3.svg.axis()
    .scale(y)
    .orient('left');

  const line = d3.svg.line()
    .x((d) => { return t(d.timestamp); })
    .y((d) => { return y(d.stars); });

  const svg = d3
    .select(STAR_COUNT_SELECTOR)
    .append('svg')
      .attr('class', 'chart__svg')
      .attr('width', width + LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right)
      .attr('height', height + LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom)
    .append('g')
      .attr('transform', `translate(${LINE_CHART_MARGIN.left},${LINE_CHART_MARGIN.top})`);
    
  t.domain(d3.extent(formattedCounts, (d) => { return d.timestamp; }));
  y.domain(d3.extent(formattedCounts, (d) => { return d.stars; }));

  svg.append('g')
    .attr('class', 'chart__title')
    .append('text')
      .attr('class', 'chart__title-text')
      .attr('transform', `translate(${width / 2 - 60}, 0)`)
      .text('Stars Over Time');

  svg.append('g')
    .attr('class', 'chart__x-axis')
    .attr('transform', `translate(0,${height})`)
    .call(tAxis);

  svg.append('g')
      .attr('class', 'chart__y-axis')
      .call(yAxis)
    .append('text')
      .attr('transform', 'rotate(-90)')
      .attr('y', 6)
      .attr('dy', '.71em')
      .style('text-anchor', 'end')
      .text('Stars');

  svg.append('path')
    .datum(formattedCounts)
    .attr('class', 'chart__line chart__line--orange')
    .attr('d', line);

  const starHeadlineTpl = document.querySelector('.template__star-headline');
  const starHeadline = starHeadlineTpl.cloneNode(true);
  starHeadline
    .querySelector('.star-headline__count')
    .appendChild(document.createTextNode(starCounts[starCounts.length-1].Stars));
  starHeadline.className = '';
  document.querySelector('.tile__star-headline')
    .appendChild(starHeadline);
});

fetch('/gh/rackerlabs/mimic/open_issue_counts').then((response) => {
  return response.json();
})
.then(function (issueCounts) {
  const formattedCounts = map(issueCounts, (issueCount) => {
    return { openIssues: issueCount.OpenIssues,
             timestamp: d3.time.format.iso.parse(issueCount.Timestamp) };
  });

  const issueCountEl = $(ISSUE_COUNT_SELECTOR);

  const height =
    issueCountEl.height() - (LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom);
  const width =
    issueCountEl.width() - (LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right);

  const t = d3.time.scale()
    .range([0, width]);

  const tAxis = d3.svg.axis()
    .scale(t)
    .orient('bottom');

  const y = d3.scale.linear()
    .range([height, 0]);

  const yAxis = d3.svg.axis()
    .scale(y)
    .orient('left');

  const line = d3.svg.line()
    .x((d) => { return t(d.timestamp); })
    .y((d) => { return y(d.openIssues); });

  const svg = d3
    .select(ISSUE_COUNT_SELECTOR)
    .append('svg')
      .attr('class', 'chart__svg')
      .attr('width', width + LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right)
      .attr('height', height + LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom)
    .append('g')
      .attr('transform', `translate(${LINE_CHART_MARGIN.left},${LINE_CHART_MARGIN.top})`);
    
  t.domain(d3.extent(formattedCounts, (d) => { return d.timestamp; }));
  y.domain(d3.extent(formattedCounts, (d) => { return d.openIssues; }));

  svg.append('g')
    .attr('class', 'chart__title')
    .append('text')
      .attr('class', 'chart__title-text')
      .attr('transform', `translate(${width / 2 - 80}, 0)`)
      .text('Open Issues Over Time');

  svg.append('g')
    .attr('class', 'chart__x-axis')
    .attr('transform', `translate(0,${height})`)
    .call(tAxis);

  svg.append('g')
      .attr('class', 'chart__y-axis')
      .call(yAxis)
    .append('text')
      .attr('transform', 'rotate(-90)')
      .attr('y', 6)
      .attr('dy', '.71em')
      .style('text-anchor', 'end')
      .text('Open Issues');

  svg.append('path')
    .datum(formattedCounts)
    .attr('class', 'chart__line chart__line--green')
    .attr('d', line);
});

fetch('/gh/rackerlabs/mimic/open_pr_counts').then((response) => {
  return response.json();
})
.then(function (prCounts) {
  const formattedCounts = map(prCounts, (prCount) => {
    return { openIssues: prCount.OpenIssues,
             timestamp: d3.time.format.iso.parse(prCount.Timestamp) };
  });

  const prCountEl = $(PR_COUNT_SELECTOR);

  const height =
    prCountEl.height() - (LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom);
  const width =
    prCountEl.width() - (LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right);

  const t = d3.time.scale()
    .range([0, width]);

  const tAxis = d3.svg.axis()
    .scale(t)
    .orient('bottom');

  const y = d3.scale.linear()
    .range([height, 0]);

  const yAxis = d3.svg.axis()
    .scale(y)
    .orient('left');

  const line = d3.svg.line()
    .x((d) => { return t(d.timestamp); })
    .y((d) => { return y(d.openIssues); });

  const svg = d3
    .select(PR_COUNT_SELECTOR)
    .append('svg')
      .attr('class', 'chart__svg')
      .attr('width', width + LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right)
      .attr('height', height + LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom)
    .append('g')
      .attr('transform', `translate(${LINE_CHART_MARGIN.left},${LINE_CHART_MARGIN.top})`);
    
  t.domain(d3.extent(formattedCounts, (d) => { return d.timestamp; }));
  y.domain(d3.extent(formattedCounts, (d) => { return d.openIssues; }));

  svg.append('g')
    .attr('class', 'chart__title')
    .append('text')
      .attr('class', 'chart__title-text')
      .attr('transform', `translate(${width / 2 - 80}, 0)`)
      .text('Open PRs Over Time');

  svg.append('g')
    .attr('class', 'chart__x-axis')
    .attr('transform', `translate(0,${height})`)
    .call(tAxis);

  svg.append('g')
      .attr('class', 'chart__y-axis')
      .call(yAxis)
    .append('text')
      .attr('transform', 'rotate(-90)')
      .attr('y', 6)
      .attr('dy', '.71em')
      .style('text-anchor', 'end')
      .text('Open PRs');

  svg.append('path')
    .datum(formattedCounts)
    .attr('class', 'chart__line chart__line--blue')
    .attr('d', line);
});
