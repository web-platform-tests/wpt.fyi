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

/**
 * Generates a font and background color for percentages on the Interop
 * dashboardbased on a score from 0 to 100. The colors are calculated on a
 * gradient from red to green.
 *
 * @param {number} score - The score, ranging from 0 to 100.
 * @returns {[string, string]} An array containing the font color as an RGB
 * string and the background color as an RGBA string with 15% opacity.
 */
export function calculateColor(score) {
  const gradient = [
    // Red.
    { scale: 0, color: [250, 0, 0] },
    // Orange.
    { scale: 33.33, color: [250, 125, 0] },
    // Yellow.
    { scale: 66.67, color: [220, 220, 0] },
    // Green.
    { scale: 100, color: [0, 160, 0] },
  ];

  let color1, color2;
  for (let i = 1; i < gradient.length; i++) {
    if (score <= gradient[i].scale) {
      color1 = gradient[i - 1];
      color2 = gradient[i];
      break;
    }
  }
  const colorWeight = ((score - color1.scale) / (color2.scale - color1.scale));
  const color = [
    Math.round(color1.color[0] * (1 - colorWeight) + color2.color[0] * colorWeight),
    Math.round(color1.color[1] * (1 - colorWeight) + color2.color[1] * colorWeight),
    Math.round(color1.color[2] * (1 - colorWeight) + color2.color[2] * colorWeight),
  ];

  return [
    `rgb(${color[0]}, ${color[1]}, ${color[2]})`,
    `rgba(${color[0]}, ${color[1]}, ${color[2]}, 0.15)`,
  ];
}

export { timeTaken };
