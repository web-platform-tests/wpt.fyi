import { assert } from '@open-wc/testing';
import {interopData} from '../interop-data.js';

// Check that the data in the JavaScript file and in the JSON file match.
// interop-data.json is used by some Mozilla infrastructure, so the file should
// not be deleted and should remain up-to-date.
suite('contents of webapp/components/interop-data.js', () => {
  test('should match webapp/static/interop-data.json exactly', async() => {
    const resp = await fetch('/static/interop-data.json');
    const interopJson = await resp.json();
    assert.deepEqual(interopData, interopJson);
  });
});
