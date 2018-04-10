# UI Development

The UI consists of base HTML templates served by the Go App Engine app in `webapp/templates/` and Polymer 2 components in `webapp/components/`.

## UI Principles

- The dashboard should not surface any overall metrics that compare complete runs of different browsers against each other.
- Clean, uncluttered design.

### More specifically

- All pages should be interactable [within 1000ms](https://developers.google.com/web/fundamentals/performance/rail#load) and fully loaded within 2000ms on a good connection.
- All fonts are over `15px`.
