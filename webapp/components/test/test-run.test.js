import { assert, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import { TestRun } from '../test-run.js';

suite('<test-run>', () => {
  let trf = null;

  setup(() => {
    trf = fixtureSync(html`<test-run></test-run>`);
    trf.testRun = {
      browser_name: 'firefox',
      time_start: '2018-01-12T12:00:00Z',
      time_end: '2018-01-12T13:20:00Z',
      revision: '0123456789',
    };
  });

  suite('static get properties()', () => {
    test('testRun', () => {
      assert.property(TestRun.properties, 'testRun');
      assert.property(TestRun.properties.testRun, 'type');
      assert.equal(TestRun.properties.testRun.type, Object);
    });
  });

  suite('TestRun.prototype.*', () => {
    let sandbox;

    setup(() => {
      sandbox = sinon.createSandbox();
      // Override the timezone as UTC.
      for (const method of ['toLocaleDateString', 'toLocaleTimeString']) {
        const original = Date.prototype[method];
        sandbox.stub(Date.prototype, method).callsFake(function(locale, options) {
          return original.call(
            this, 'en-US', Object.assign(options, {timeZone: 'UTC'}));
        });
      }
    });

    suite('dateFormat()', () => {
      test('dateFormat(iso string)', () => {
        assert.equal(trf.dateFormat('2018-01-12T12:00:00Z'), 'Jan 12, 2018');
      });
    });

    test('timeFormat(iso string)', () => {
      assert.equal(trf.timeFormat('2018-01-12T12:00:00Z'), 'Jan 12 at 12:00 PM');
    });

    suite('timeTaken(testRun Object)', () => {
      test('1h 20m', () => {
        const testRun = {
          time_start: '2018-01-12T12:00:00Z',
          time_end: '2018-01-12T13:20:00Z',
        };
        assert.equal(trf.timeTaken(testRun), '(took 1.3h)');
      });

      test('0m', () => {
        const testRun = {
          time_start: '2018-01-12T12:00:00Z',
          time_end: '2018-01-12T12:00:00Z',
        };
        assert.equal(trf.timeTaken(testRun), '');
      });
    });

    suite('isDiff()', () => {
      test('isDiff("Diff|diff")', () => {
        assert.isTrue(trf.isDiff('Diff'));
        assert.isTrue(trf.isDiff('diff'));
      });
      test('isDiff(anything else)', () => {
        assert.isFalse(trf.isDiff('literally anything'));
      });
    });

    suite('shortVersion', () => {
      test('valid, major only', () => {
        assert.equal(trf.shortVersion('chrome', '1'), '1');
        assert.equal(trf.shortVersion('chrome', '2.3'), '2');
        assert.equal(trf.shortVersion('chrome', '3.4.5'), '3');
        assert.equal(trf.shortVersion('chrome', '4.5.6.7'), '4');
        assert.equal(trf.shortVersion('chrome', '765.687'), '765');
        assert.equal(trf.shortVersion('chrome', '   11.0 '), '11');
        assert.equal(trf.shortVersion('chrome', '56.0a'), '56');
      });

      test('valid, major and minor', () => {
        assert.equal(trf.shortVersion('safari', '1'), '1');
        assert.equal(trf.shortVersion('safari', '2.3'), '2.3');
        assert.equal(trf.shortVersion('safari', '3.4.5'), '3.4');
        assert.equal(trf.shortVersion('safari', '4.5.6.7'), '4.5');
        assert.equal(trf.shortVersion('safari', '765.687'), '765.687');
        assert.equal(trf.shortVersion('safari', '   11.0 '), '11.0');
        assert.equal(trf.shortVersion('safari', '56.0a'), '56.0');
      });

      test('valid, major and minor', () => {
        assert.equal(trf.shortVersion('webkitgtk', '1'), '1');
        assert.equal(trf.shortVersion('webkitgtk', '2.3'), '2.3');
        assert.equal(trf.shortVersion('webkitgtk', '3.4.5'), '3.4');
        assert.equal(trf.shortVersion('webkitgtk', '4.5.6.7'), '4.5');
        assert.equal(trf.shortVersion('webkitgtk', '765.687'), '765.687');
        assert.equal(trf.shortVersion('webkitgtk', '   11.0 '), '11.0');
        assert.equal(trf.shortVersion('webkitgtk', '56.0a'), '56.0');
      });

      test('valid, major and minor', () => {
        assert.equal(trf.shortVersion('wpewebkit', '1'), '1');
        assert.equal(trf.shortVersion('wpewebkit', '2.3'), '2.3');
        assert.equal(trf.shortVersion('wpewebkit', '3.4.5'), '3.4');
        assert.equal(trf.shortVersion('wpewebkit', '4.5.6.7'), '4.5');
        assert.equal(trf.shortVersion('wpewebkit', '765.687'), '765.687');
        assert.equal(trf.shortVersion('wpewebkit', '   11.0 '), '11.0');
        assert.equal(trf.shortVersion('wpewebkit', '56.0a'), '56.0');
      });

      test('invalid', () => {
        assert.equal(trf.shortVersion('chrome', 'five'), 'five');
        assert.equal(trf.shortVersion('chrome', ''), '');
        assert.equal(trf.shortVersion('safari', 'five'), 'five');
        assert.equal(trf.shortVersion('safari', ''), '');
        assert.equal(trf.shortVersion('webkitgtk', 'five'), 'five');
        assert.equal(trf.shortVersion('webkitgtk', ''), '');
        assert.equal(trf.shortVersion('wpewebkit', 'five'), 'five');
        assert.equal(trf.shortVersion('wpewebkit', ''), '');
      });
    });

    teardown(() => {
      sandbox.restore();
    });
  });
});
