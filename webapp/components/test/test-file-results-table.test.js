import { assert, fixtureSync, html } from '@open-wc/testing';
import { TestFileResultsTable } from '../test-file-results-table.js';

suite('TestFileResultsTable', () => {
  let tfrt;

  setup(() => {
    tfrt = fixtureSync(html`
      <test-file-results-table
        path="/2dcontext/the-canvas-state/2d.state.saverestore.bitmap.html">
      </test-file-results-table>
    `);
  });

  test('is a registered custom element', () => {
    assert.isTrue(new TestFileResultsTable() instanceof HTMLElement);
    assert.isTrue(document.createElement('test-file-results-table') instanceof TestFileResultsTable);
  });

  suite('static get is()', () => {
    test('test-file-results-table', () => {
      assert.equal(TestFileResultsTable.is, 'test-file-results-table');
    });
  });

  suite('TestFileResultsTable.prototype.*', () => {
    suite('colorClass', () => {
      test('Pass and fail are colored differently', () => {
        assert.notEqual(tfrt.colorClass('PASS'), tfrt.colorClass('FAIL'));
      });
      test('Fail and error are colored identically', () => {
        assert.equal(tfrt.colorClass('FAIL'), tfrt.colorClass('ERROR'));
      });
      test('Harness status OK is equivalent to pass', () => {
        // OK is used for harness status and should not be colored green.
        assert.notEqual(tfrt.colorClass('OK'), tfrt.colorClass('PASS'));
      });
    });

    suite('computeDisplayedProducts', () => {
      test('null testRuns', () => {
        assert.deepEqual(tfrt.computeDisplayedProducts(null), []);
      });
      test('simple testRuns', () => {
        const testRuns = [{
          browser_name: 'firefox',
          browser_version: 1,
          labels: ['labelA'],
          revision: '0123456789',
        }];

        const result = tfrt.computeDisplayedProducts(testRuns);

        assert.equal(result.length, 1);
        assert.equal(result[0].browser_name, 'firefox');
        assert.equal(result[0].browser_version, 1);
        assert.deepEqual(result[0].labels, ['labelA']);
        assert.equal(result[0].revision, '0123456789');
      });
    });
  });

});
