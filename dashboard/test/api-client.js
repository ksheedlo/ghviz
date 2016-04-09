import ApiClient from '../api-client';
import Cache from '../cache';
import chai, { expect } from 'chai';
import sinonChai from 'sinon-chai';
chai.use(sinonChai);

describe('ApiClient', () => {
  let apiClient;

  beforeEach(() => {
    sinon.stub(window, 'fetch');
    apiClient = new ApiClient({ cache: new Cache({ maxAge: 1000 * 60 * 5 }) });
  });

  afterEach(() => {
    window.fetch.restore();
  });

  describe('.listStarCounts', () => {
    it('fetches the star counts', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        '[{"stars":1,"timestamp":"2016-03-25T19:46:41.395Z"}]',
        { status: 200,
          headers: { 'Content-Type': 'application/json' } })));

      apiClient.listStarCounts({
        owner: 'tester',
        repo: 'cool-project',
      })
      .then((counts) => {
        expect(window.fetch)
          .to.have.been.calledWith('/gh/tester/cool-project/star_counts');
        expect(counts).to.eql([{
          stars: 1,
          timestamp: '2016-03-25T19:46:41.395Z',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });

    it('caches the star counts on success', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        '[{"stars":1,"timestamp":"2016-03-25T19:46:41.395Z"}]',
        { status: 200,
          headers: { 'Content-Type': 'application/json' } })));
      apiClient.listStarCounts({
        owner: 'tester',
        repo: 'cool-project',
      }).then(() => {
        return apiClient.listStarCounts({ owner: 'tester',
                                          repo: 'cool-project' });
      })
      .then((counts) => {
        expect(window.fetch.callCount).to.equal(1);
        expect(counts).to.eql([{
          stars: 1,
          timestamp: '2016-03-25T19:46:41.395Z',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });

    it('refetches if an error occurs', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        'Server Error',
        { status: 500,
          headers: { 'Content-Type': 'text/plain' } })));
      apiClient.listStarCounts({
        owner: 'tester',
        repo: 'cool-project',
      })
      .catch(() => {
        window.fetch.returns(Promise.resolve(new window.Response(
          '[{"stars":1,"timestamp":"2016-03-25T19:46:41.395Z"}]',
          { status: 200,
            headers: { 'Content-Type': 'application/json' } })));

        return apiClient.listStarCounts({ owner: 'tester',
                                          repo: 'cool-project' });
      })
      .then((counts) => {
        expect(window.fetch.callCount).to.equal(2);
        expect(counts).to.eql([{
          stars: 1,
          timestamp: '2016-03-25T19:46:41.395Z',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });
  });

  describe('.listIssueCounts', () => {
    it('fetches the star counts', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        '[{"open_issues":1,"timestamp":"2016-03-25T19:46:41.395Z"}]',
        { status: 200,
          headers: { 'Content-Type': 'application/json' } })));

      apiClient.listIssueCounts({
        owner: 'tester',
        repo: 'cool-project',
      })
      .then((counts) => {
        expect(window.fetch)
          .to.have.been.calledWith('/gh/tester/cool-project/issue_counts');
        expect(counts).to.eql([{
          open_issues: 1,
          timestamp: '2016-03-25T19:46:41.395Z',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });

    it('caches the star counts on success', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        '[{"open_issues":1,"timestamp":"2016-03-25T19:46:41.395Z"}]',
        { status: 200,
          headers: { 'Content-Type': 'application/json' } })));

      apiClient.listIssueCounts({
        owner: 'tester',
        repo: 'cool-project',
      }).then(() => {
        return apiClient.listIssueCounts({ owner: 'tester',
                                           repo: 'cool-project' });
      })
      .then((counts) => {
        expect(window.fetch.callCount).to.equal(1);
        expect(counts).to.eql([{
          open_issues: 1,
          timestamp: '2016-03-25T19:46:41.395Z',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });

    it('refetches if an error occurs', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        'Server Error',
        { status: 500,
          headers: { 'Content-Type': 'text/plain' } })));
      apiClient.listIssueCounts({
        owner: 'tester',
        repo: 'cool-project',
      })
      .catch(() => {
        window.fetch.returns(Promise.resolve(new window.Response(
          '[{"open_issues":1,"timestamp":"2016-03-25T19:46:41.395Z"}]',
          { status: 200,
            headers: { 'Content-Type': 'application/json' } })));

        return apiClient.listIssueCounts({ owner: 'tester',
                                           repo: 'cool-project' });
      })
      .then((counts) => {
        expect(window.fetch.callCount).to.equal(2);
        expect(counts).to.eql([{
          open_issues: 1,
          timestamp: '2016-03-25T19:46:41.395Z',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });
  });

  describe('.listTopIssues', () => {
    it('fetches the top issues', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        '[{"title":"Test Issue"}]',
        { status: 200,
          headers: { 'Content-Type': 'application/json' } })));

      apiClient.listTopIssues({
        owner: 'tester',
        repo: 'cool-project',
      })
      .then((topIssues) => {
        expect(window.fetch)
          .to.have.been.calledWith('/gh/tester/cool-project/top_issues');
        expect(topIssues).to.eql([{
          title: 'Test Issue',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });
  });

  describe('.listTopPrs', () => {
    it('fetches the top PRs', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        '[{"title":"Test PR"}]',
        { status: 200,
          headers: { 'Content-Type': 'application/json' } })));

      apiClient.listTopPrs({
        owner: 'tester',
        repo: 'cool-project',
      })
      .then((topPrs) => {
        expect(window.fetch)
          .to.have.been.calledWith('/gh/tester/cool-project/top_prs');
        expect(topPrs).to.eql([{
          title: 'Test PR',
        }]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });
  });

  describe('.listTopContributors', () => {
    it('lists the monthly top contributors', (done) => {
      window.fetch.returns(Promise.resolve(new window.Response(
        '[{"name":"tester1"}, {"name":"tester2"}]',
        { status: 200,
          headers: { 'Content-Type': 'application/json' } })));

      apiClient.listTopContributors({
        date: new Date(1458940112436),
        owner: 'tester',
        repo: 'cool-project',
      })
      .then((topContributors) => {
        expect(window.fetch)
          .to.have.been.calledWith('/gh/tester/cool-project/highscores/2016/03');
        expect(topContributors).to.eql([
          { name: 'tester1' },
          { name: 'tester2' },
        ]);
        done();
      })
      .catch((err) => {
        throw err;
      });
    });
  });
});
