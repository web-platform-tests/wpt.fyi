function timeTaken(millis) {
  if (!millis) {
    return 'N/A';
  }
  const s = millis / 1000;
  const units = [
    [60 * 60 * 24 * 365, 'years'],
    [60 * 60 * 24 * 28, 'months'],
    [60 * 60 * 24 * 7, 'weeks'],
    [60 * 60 * 24, 'days'],
    [60 * 60, 'hours'],
    [60, 'minutes'],
  ];
  for (const unit of units) {
    const scalar = Math.floor(s / unit[0]);
    if (scalar > 1) {
      return `${scalar} ${unit[1]}`;
    }
  }
  return `${s} seconds`;
}

export { timeTaken };
