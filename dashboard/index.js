'use strict';

const bluebird = require('bluebird');
if (!window.Promise) {
  window.Promise = bluebird;
}
require('whatwg-fetch');

const React = require('react');
const ReactDOM = require('react-dom');
const TopIssues = require('./components/TopIssues');
const TopPrs = require('./components/TopPrs');

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
  forEach = require('lodash.foreach'),
  map = require('lodash.map'),
  $ = require('jquery');

const owner = window.GLOBALS.owner,
  repo = window.GLOBALS.repo;

function starCaption(stars) {
  if (stars < 10) {
    return 'Imagine the possibilities!';
  }
  if (stars < 100) {
    return 'This is starting to pick up steam!';
  }
  if (stars < 1000) {
    return 'It\'ll take over the world someday.';
  }
  if (stars < 9001) {
    return 'Look at all the Internet points!';
  }
  if (stars < 10000) {
    return 'IT\'S OVER NINE THOUSAND!';
  }
  if (stars < 100000) {
    return 'Literally bigger than jQuery.';
  }
  return 'World Domination';
}

fetch(`/gh/${owner}/${repo}/star_counts`).then((resp) => {
  return resp.json();
})
.then((starCounts) => {
  const formattedCounts = map(starCounts, (starCount) => {
    return { stars: starCount.stars,
             timestamp: d3.time.format.iso.parse(starCount.timestamp) };
  });

  const starCountQsel = document.querySelector('.tile__star-count');
  starCountQsel.removeChild(starCountQsel.querySelector('.loader__wrapper'));

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

  const path = svg.append('path')
    .datum(formattedCounts)
    .attr('class', 'chart__line chart__line--orange')
    .attr('d', line);

  const totalLength = path.node().getTotalLength();

  path
    .attr("stroke-dasharray", totalLength + " " + totalLength)
    .attr("stroke-dashoffset", totalLength)
    .transition()
      .duration(1000)
      .ease("linear")
      .attr("stroke-dashoffset", 0);

  const starHeadlineTile = document.querySelector('.tile__star-headline');
  starHeadlineTile.removeChild(starHeadlineTile.querySelector('.loader__wrapper'));
  const starHeadlineTpl = document.querySelector('.template__star-headline');
  const starHeadline = starHeadlineTpl.cloneNode(true);
  starHeadline
    .querySelector('.star-headline__count')
    .appendChild(document.createTextNode(starCounts[starCounts.length-1].stars));
  starHeadline
    .querySelector('.star-headline__caption')
    .appendChild(document.createTextNode(starCaption(
      starCounts[starCounts.length-1].stars)));
  starHeadline.className = '';
  starHeadlineTile.appendChild(starHeadline);
});

function drawIssues({ chartLineColor,
                      issueCounts,
                      issueCountSelector,
                      key,
                      title,
                      yLabel }) {
  const issueCountQsel = document.querySelector(issueCountSelector);
  issueCountQsel.removeChild(issueCountQsel.querySelector('.loader__wrapper'));

  const issueCountEl = $(issueCountSelector);

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
    .y((d) => { return y(d[key]); });

  const svg = d3
    .select(issueCountSelector)
    .append('svg')
      .attr('class', 'chart__svg')
      .attr('width', width + LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right)
      .attr('height', height + LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom)
    .append('g')
      .attr('transform', `translate(${LINE_CHART_MARGIN.left},${LINE_CHART_MARGIN.top})`);
    
  t.domain(d3.extent(issueCounts, (d) => { return d.timestamp; }));
  y.domain(d3.extent(issueCounts, (d) => { return d[key]; }));

  svg.append('g')
    .attr('class', 'chart__title')
    .append('text')
      .attr('class', 'chart__title-text')
      .attr('transform', `translate(${width / 2 - 80}, 0)`)
      .text(title);

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
      .text(yLabel);

  const path = svg.append('path')
    .datum(issueCounts)
    .attr('class', `chart__line chart__line--${chartLineColor}`)
    .attr('d', line);

  const totalLength = path.node().getTotalLength();

  path
    .attr("stroke-dasharray", totalLength + " " + totalLength)
    .attr("stroke-dashoffset", totalLength)
    .transition()
      .duration(1000)
      .ease("linear")
      .attr("stroke-dashoffset", 0);
}

const issueCountsPromise = fetch(`/gh/${owner}/${repo}/issue_counts`).then(
  (response) => {
  return response.json();
});

issueCountsPromise.then((issueCounts) => {
  const formattedCounts = map(issueCounts, (issueCount) => {
    return { openIssues: issueCount.open_issues,
             openPrs: issueCount.open_prs,
             timestamp: d3.time.format.iso.parse(issueCount.timestamp) };
  });

  drawIssues({
    chartLineColor: 'green',
    issueCounts: formattedCounts,
    issueCountSelector: ISSUE_COUNT_SELECTOR,
    key: 'openIssues',
    title: 'Open Issues Over Time',
    yLabel: 'Open Issues'
  });

  drawIssues({
    chartLineColor: 'blue',
    issueCounts: formattedCounts,
    issueCountSelector: PR_COUNT_SELECTOR,
    key: 'openPrs',
    title: 'Open PRs Over Time',
    yLabel: 'Open PRs'
  });
});

ReactDOM.render(
  <TopIssues />,
  document.querySelector('.holder__top-issues')
);

ReactDOM.render(
  <TopPrs />,
  document.querySelector('.holder__top-prs')
);
