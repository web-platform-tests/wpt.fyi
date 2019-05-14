const log = require('debug')('wpt.fyi');
const {Datastore} = require('@google-cloud/datastore');

async function populate(url) {
  const datastore = new Datastore({
    projectId: 'wptdashboard',
    apiEndpoint: url,
  });

  const staticRunSHA = 'b952881825e7d3974f5c513e13e544d525c0a631';
  const key = datastore.key('Cat');
  log(key)
  const testRun = {
    key: key,
    data: {
      browser_name: 'chrome',
      full_revision_hash: staticRunSHA,
      revision:           staticRunSHA.slice(0, 10),
    }
  };
  const foo = await datastore.save(testRun);
  log(`Added TestRun ${testRun.data[datastore.KEY]}`);
}

populate('http://localhost:8081/');