class TestStatus {
  constructor(name) {
    this.name = name;
  }

  get isPass() {
    return this.name === 'PASS' || this.name === 'OK';
  }

  toString() {
    return this.name;
  }
}

const TestStatuses = Object.freeze({
  UNKNOWN: new TestStatus('UNKNOWN'),
  PASS: new TestStatus('PASS'),
  OK: new TestStatus('OK'),
  ERROR: new TestStatus('ERROR'),
  TIMEOUT: new TestStatus('TIMEOUT'),
  NOTRUN: new TestStatus('NOTRUN'),
  FAIL: new TestStatus('FAIL'),
  CRASH: new TestStatus('CRASH'),
  SKIP: new TestStatus('SKIP'),
  ASSERT: new TestStatus('ASSERT'),
});

export { TestStatuses };
