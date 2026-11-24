# SocialPredict product tour

A nine-slide static presentation at https://openpredictionmarkets.github.io/socialpredict/.
The repository's top-level README links to that URL. The Pages workflow deploys
the site root from this directory after a merge into main, with no frontend or
backend build required. Pull requests validate the tour without deploying.

Preview from the repository root:

```sh
python3 -m http.server 4173 --directory docs/tour
```

Open http://localhost:4173/. Use Previous, Next, the chapter picker, or the left
and right arrow keys. Home and End jump to the first and last slide. Each slide
has a hash link. Browser Back and Forward work. Reset tour or reload clears
sample changes. Arrow keys retain native behavior in inputs and controls.

Verify:

```sh
cd docs/tour
npm ci
npm test
npx playwright install chromium
npm run test:browser
```

The browser test serves the tour under `/socialpredict/`, exercises the main
flows, checks every original capture, and checks all slides at 375, 768, and
1440 pixels. Runtime files have no package dependencies or network calls.

## Reuse and source material

- Brand palette: `frontend/tailwind.config.js`; logo:
  `frontend/src/assets/logo/socialpredictlogo.svg`.
- Probability formula: `weightedAverageProbabilityFormula.Calculate` in
  `backend/internal/domain/math/probabilities/wpam/wpam_marketprobabilities.go`.
  Demo seeds and fee match `backend/setup/setup.yaml`: initial probability 0.5,
  subsidy 10, initial YES/NO contributions 0, first trade fee 1.
- Trade preview flow: `frontend/src/components/layouts/trade/BuySharesLayout.jsx`
  and `frontend/src/components/layouts/marketprojection/MarketProjectionLayout.jsx`.
- Original screenshots and captions: the user's supplied Google Docs appendices
  A (admin platform), B (moderator and topics), C (participant search and trading),
  November 23, 2026. `sources.mjs` records all 41 images, captions, verification
  notes and tab links. The source snapshot is
  `d133d22751b21f38fa54acbfccb6db98e719fa9a`.

All interactive controls manipulate in-memory examples. They do not create
accounts, change platform settings, or call the SocialPredict API. The sale
replay uses recorded fixtures rather than implementing the production share
allocation and sale algorithms. Its five-credit sale and final 33 YES shares
describe that market history only. Group sliders illustrate independent outcomes;
they do not represent executed trades. The walkthrough resolved no question.

The evidence viewer retains the appendices' distinction between exercised and
captured-only controls. Suspension, tag deactivation, image uploads and crawler
rendering are not presented as verified actions. Topic links and scoped search
were exercised in B/C; optional chart pins stayed disabled after the documented
display mismatch. No private grant narrative or real credentials ship here.

## Pages settings

The repository must use **Settings → Pages → Source → GitHub Actions**. The
workflow packages only the static runtime files and screenshot assets. It does
not upload node_modules, tests, this guide, or other repository documentation.
Deployment runs only on main; no pull-request code receives Pages write access.
