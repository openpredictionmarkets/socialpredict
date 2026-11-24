/* Original appendix captures, November 23, 2026; captions retained for provenance. */
export const sources = [
  {
    "id": "A1",
    "title": "An organizer’s own platform",
    "caption": "The saved homepage presents Workshop Forecast Hub’s purpose, joining instructions, and community rules. This is platform-wide content around a catalogue of markets, with separate participant accounts and moderator governance. The SocialPredict navigation brand remains in this build; changing that shell is a deployment/source customization.",
    "verification": "Verified: homepage saved through the CMS and displayed on the running site.",
    "image": "assets/screenshots/admin-platform/09-custom-platform-home.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A2",
    "title": "Update content without a release",
    "caption": "The Home Page editor lets an administrator change the title and HTML, save it, and inspect a live preview. An organizer can revise onboarding, house rules, and announcements without deploying a new frontend. The version field supports guarded updates; authoring HTML still requires basic markup knowledge.",
    "verification": "Verified: custom title and homepage content persisted.",
    "image": "assets/screenshots/admin-platform/08-homepage-editor.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A3",
    "title": "Create participant accounts quickly",
    "caption": "Add User creates an account and generates its initial password. The resulting credentials card supports copying login details, and Add Another User resets the form for the next participant. The initial password is masked in this screenshot. First-login password changes keep generated credentials temporary.",
    "verification": "Verified: synthetic account created and generated-credentials card displayed; password changes also exercised for admin and sample hosts.",
    "image": "assets/screenshots/admin-platform/07-account-onboarding.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A4",
    "title": "Approve who can create markets",
    "caption": "The Non-Moderators queue separates ordinary participants from approved market creators. Search and pagination make the roster manageable; an administrator can promote a selected participant to moderator and record a reason. This lets a small team recruit trusted hosts while retaining the platform’s publishing policy.",
    "verification": "Verified: two synthetic hosts promoted through the backend; promotion controls captured in the running UI.",
    "image": "assets/screenshots/admin-platform/05-user-governance.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A5",
    "title": "Delegate work with revocable authority",
    "caption": "The Moderators view exposes active status, reason fields, and suspension controls. A platform owner can distribute question creation and stewardship among trusted people, then suspend or reinstate a moderator when needed. This is moderator governance, not a demonstrated general-purpose user-ban system.",
    "verification": "Captured: live active-moderator roster and suspension controls. Suspension was not exercised.",
    "image": "assets/screenshots/admin-platform/06-moderator-controls.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A6",
    "title": "Review binary and grouped proposals",
    "caption": "The review table brings title, criteria, creator, steward, proposal cost, tags, and approval actions together. Binary questions can be approved individually; grouped questions can be approved as a unit. Proposed questions remain outside public trading until the publishing decision. Search, status tabs, and pagination keep review work in one place.",
    "verification": "Verified: one binary proposal and one two-answer group approved from this UI.",
    "image": "assets/screenshots/admin-platform/02-proposal-review.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A7",
    "title": "Reject unclear questions and refund credits",
    "caption": "Reject and Refund requires a reason. The rejected view retains the decision trail and proposal-credit refund information, giving the proposer an actionable explanation and avoiding a separate manual credit adjustment. Here the administrator requested a publication URL and an unambiguous deadline.",
    "verification": "Verified: synthetic proposal rejected; retained reason and refund trail read back. Credits are play credits.",
    "image": "assets/screenshots/admin-platform/04-rejection-refund-trail.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A8",
    "title": "Automate routine publishing",
    "caption": "The platform-wide Auto-approve new market proposals switch lets trusted moderator proposals go directly to published, tradable status. Teams can keep manual review during onboarding, then reduce recurring review work when their contributors and question format are reliable. Changing the switch does not itself prove every other workflow is automated.",
    "verification": "Verified: enabled and saved; a newly submitted moderator proposal returned lifecycleStatus=published without a separate approval call.",
    "image": "assets/screenshots/admin-platform/18-market-auto-approval.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A9",
    "title": "Review contract clarifications",
    "caption": "Description Amendments are append-only changes to published question criteria. Reviewers see the original description, proposed addition, author, and submit reason, and must supply a decision reason. The separate auto-approval switch offers a lower-touch path for new steward clarifications when that policy is appropriate.",
    "verification": "Verified: a steward submitted a clarification and an administrator approved it. Amendment auto-approval was available but not enabled in this walkthrough.",
    "image": "assets/screenshots/admin-platform/13-amendment-governance.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A10",
    "title": "Keep the clarification history",
    "caption": "Approved Amendments preserves the accepted addition and decision metadata instead of silently replacing the original contract. A small team can explain how criteria evolved and resolve participant questions from a shared record. This reduces reliance on private messages or a host’s memory.",
    "verification": "Verified: approved clarification retrieved in the history view after review.",
    "image": "assets/screenshots/admin-platform/19-amendment-history.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A11",
    "title": "Choose an answer-approval policy",
    "caption": "Grouped markets offer three platform policies: auto-approve every active moderator’s answer, let the steward choose per market, or require administrator review. This makes delegation configurable across the whole platform while retaining a stricter option for sensitive questions.",
    "verification": "Verified: saved the auto-approve policy; a subsequent answer returned status=approved and reviewedBy=auto-approval, with a new child market.",
    "image": "assets/screenshots/admin-platform/14-answer-automation-policy.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A12",
    "title": "Expand grouped markets through a queue",
    "caption": "Answer Additions records who proposed an option, its configured credit cost, and its approval outcome. Approved answers create new binary child markets while preserving existing child-market history. Administrators can review a new option without rebuilding the grouped question.",
    "verification": "Verified: manually approved the synthetic AI evaluation answer; a later Public health forecasting answer was automatically approved under the saved policy.",
    "image": "assets/screenshots/admin-platform/15-answer-review.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A13",
    "title": "Hand responsibility to another host",
    "caption": "Stewardship separates the question’s original creator from its current operational owner. An administrator chooses another active moderator, supplies a reason, and reassigns responsibility. The visible audit trail records who changed ownership and why, allowing holiday coverage and continuity without shared logins.",
    "verification": "Verified: market stewardship transferred from hostalice to hostben, retaining creator and reassignment reason.",
    "image": "assets/screenshots/admin-platform/16-stewardship-handoff.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A14",
    "title": "Maintain a shared topic vocabulary",
    "caption": "Admins define reusable market tags with names, descriptions, color keys, and ordering. Moderators can attach active tags to questions; administrators can adjust assignments during review. Deactivation prevents new assignments while retaining historical tags and pins, keeping old links and market context useful.",
    "verification": "Verified: Workshop tag created. Deactivate/reactivate controls captured but not exercised.",
    "image": "assets/screenshots/admin-platform/17-topic-taxonomy.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A15",
    "title": "Curate the whole market catalogue",
    "caption": "The discovery editor configures the main Markets page and topic pages, including page copy, topic pins, market pins, and recommendation limits. Search remains the main discovery path, with fallback lists available when curated blocks are empty. These are platform-level editorial controls a small team can maintain centrally.",
    "verification": "Captured: live layout, pin, and preview controls. Layout persistence and pin selection were not exercised.",
    "image": "assets/screenshots/admin-platform/10-discovery-layout.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A16",
    "title": "Set public reporting access centrally",
    "caption": "Two independent controls determine whether system metrics and the global leaderboard are visible to logged-out visitors or require login. The platform owner can choose the audience for aggregate reporting without modifying each market or releasing new code. The screen explains the effective visibility.",
    "verification": "Verified: saved system metrics as requiring login; global leaderboard remained public.",
    "image": "assets/screenshots/admin-platform/11-reporting-visibility.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A17",
    "title": "Reuse sharing defaults across markets",
    "caption": "Site name, description, image URL or upload, image alt text, and a share preview provide centrally managed defaults. Admins can disable image metadata when social crawlers create excess image traffic, or use a hosted image URL. One setting serves market links across the platform.",
    "verification": "Verified: custom site name and description saved. Image upload/CDN delivery and external social-crawler rendering were not tested.",
    "image": "assets/screenshots/admin-platform/12-social-share.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "A18",
    "title": "Publish a shared forecasting destination",
    "caption": "The public catalogue exposes the published binary and grouped questions from this instance. Participant-facing market discovery is the result of the organizer’s account, content, governance, and publishing choices working together. The demo uses synthetic questions and play credits; it is not evidence of adoption or reduced staffing cost.",
    "verification": "Verified: published questions displayed on the running platform after manual and automatic publishing.",
    "image": "assets/screenshots/admin-platform/20-public-market-platform.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.3sqetzgfjt4c"
  },
  {
    "id": "B1",
    "title": "Start with a shared topic vocabulary",
    "caption": "Admin handoff: define Useful Forecasts, Research & Replication, and Community Experiments as active tags. Names serve readers; stable slugs serve routes; descriptions explain what belongs in each category. This vocabulary can be reused by many moderators rather than recreated for every question.",
    "verification": "Verified: all three tags created and returned by the public tag catalogue.",
    "image": "assets/screenshots/moderator-workflow/B01-topic-vocabulary.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B2",
    "title": "Write a question connected to a decision",
    "caption": "Moderator: choose Binary Market, enter the title and measurable resolution criteria, and explain the decision the forecast could inform. This example asks whether three meetup forecasts will be cited in recorded decisions. The form displays the 10-credit proposal cost and rejection-refund rule before submission.",
    "verification": "Verified: created through the moderator UI as hostalice, not through an administrator’s account.",
    "image": "assets/screenshots/moderator-workflow/B02-create-question.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B3",
    "title": "Tag the question and set its trading close",
    "caption": "Moderator: select Useful Forecasts and Community Experiments, then set the Resolution Date Time. A question can have up to five tags, so one market can appear in more than one topic. The date picker uses local time; contract text separately specifies the evidence cutoff. Align trading close with when the outcome becomes knowable before a real launch.",
    "verification": "Verified: both selected tags persisted on the created market. Dates and evidence sources are synthetic.",
    "image": "assets/screenshots/moderator-workflow/B03-tags-and-deadline.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B4",
    "title": "Track the submission in the moderator queue",
    "caption": "After creation, the moderator’s profile opens its Proposed markets view. The new question, criteria, topic chips, deducted proposal cost, and Awaiting admin review status are visible together. The moderator can follow the proposal without access to the administrator’s review console.",
    "verification": "Verified: market #10 was created as proposed while platform auto-publishing was disabled.",
    "image": "assets/screenshots/moderator-workflow/B04-proposed-market.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B5",
    "title": "Create a set of independently testable experiments",
    "caption": "Moderator: choose Multiple-Choice Binary Group, provide answer options and criteria for each, and attach the same topics. Here the options compare decision-linked forecasts, weekly meetup questions, and a replication forecast round. Each option becomes a YES/NO child market; several options may resolve YES, so their probabilities need not sum to 100%.",
    "verification": "Verified: the UI created group #2 with three child markets and two topic tags.",
    "image": "assets/screenshots/moderator-workflow/B05-grouped-experiments.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B6",
    "title": "Hand the tagged proposal to an administrator",
    "caption": "Admin handoff: inspect the proposed question, steward, cost and topic assignments, then approve the binary question and the grouped question. Tag editing is part of the publishing review, so an administrator can correct category placement before a question reaches its audience.",
    "verification": "Verified: the binary and grouped proposals were approved through the review UI.",
    "image": "assets/screenshots/moderator-workflow/B08-tagged-proposal-review.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B7",
    "title": "Maintain the published question as its steward",
    "caption": "Moderator: open the published market to see its probability history, topic chips, and append-only description-amendment form. The moderator’s navigation includes Create and My Market Reviews. Steward-specific controls are separate from the administrator’s platform settings; the participant view in Appendix C omits these authoring controls.",
    "verification": "Verified: captured in a fresh moderator-only browser context. The steward’s amendment form is visible.",
    "image": "assets/screenshots/moderator-workflow/B09-moderator-market-view.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B8",
    "title": "Enable topics on the catalogue",
    "caption": "Admin handoff: open CMS, Market Discovery Layout, and the main Markets page. Set its title and description, enable Topic Pins, and save. This chooses a shared navigation structure for the whole catalogue; selecting a tag on a question alone does not automatically add a navigation button.",
    "verification": "Verified: topic navigation enabled and saved. The final walkthrough uses topic links and market tables, with optional chart pins disabled.",
    "image": "assets/screenshots/moderator-workflow/B06-enable-topic-navigation.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B9",
    "title": "Turn tag slugs into navigation links",
    "caption": "Admin handoff: add a Topic Pin for each category, set its visible Nav Label, choose the Target Topic Tag, order the entries, and save. Useful Forecasts points to useful-forecasts; Research & Replication to research-replication; Community Experiments to community-experiments. The saved bar is shared by the main catalogue and topic pages.",
    "verification": "Verified: three ordered topic pins persisted. Participant clicks to the research-replication route are shown in Appendix C.",
    "image": "assets/screenshots/moderator-workflow/B07-topic-pin-routing.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "B10",
    "title": "Give a topic its own explanation",
    "caption": "Admin handoff: select the Research & Replication secondary page and save its title and description. Its search scope is fixed to that tag, while its navigation inherits the main page’s topic links. A tagged question therefore becomes reachable through both its topic and the all-market catalogue without duplicating the market.",
    "verification": "Verified: saved page has primaryTagSlug=research-replication and searchScope=tag; participant navigation and search were exercised.",
    "image": "assets/screenshots/moderator-workflow/B11-topic-page-settings.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.stbpowk1kp38"
  },
  {
    "id": "C1",
    "title": "Enter through topics or the full catalogue",
    "caption": "Participant: open Markets. Useful Forecasts, Research & Replication, and Community Experiments are now navigation links above the market table. The ordinary account has Home, Profile, Markets, About and Stats, without moderator creation or administrator CMS links. The table retains probability, topic chips, closing date and activity context.",
    "verification": "Verified: captured as regular user learnercara after saving the topic navigation.",
    "image": "assets/screenshots/participant-workflow/C01-catalogue-and-topics.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C2",
    "title": "Follow a topic to its tagged questions",
    "caption": "Participant: click Research & Replication. The route changes to /markets/topic/research-replication, the page shows its own explanatory copy, and the table contains the independent-replication question carrying that tag. The shared topic bar lets readers move between categories without returning to a separate directory.",
    "verification": "Verified: topic link clicked and tag-filtered page loaded; the question also carries Useful Forecasts, so it can belong to both topics.",
    "image": "assets/screenshots/participant-workflow/C02-topic-navigation.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C3",
    "title": "Search across the catalogue",
    "caption": "Participant: return to Markets and type meetup. Search results include the decision-use binary question and the grouped experiment question. Active, Closed, Resolved and All determine the lifecycle scope. Searching from the main catalogue searches across topics; clear the query to return to the catalogue.",
    "verification": "Verified: global search executed through the UI. A grouped question is displayed as a grouped result rather than one unrelated row for every answer.",
    "image": "assets/screenshots/participant-workflow/C03b-global-search.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C4",
    "title": "Search within one topic",
    "caption": "Participant: return to Research & Replication and type independent. The interface explicitly says it is searching active markets tagged research-replication. This combines a navigation choice with text search, making tags useful both for browsing and for narrowing a larger catalogue.",
    "verification": "Verified: UI search and backend readback used tagSlug=research-replication and returned the tagged replication question.",
    "image": "assets/screenshots/participant-workflow/C03-scoped-search.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C5",
    "title": "Read what will count before trading",
    "caption": "Participant: open the decision-use market and expand Show Full Contract Text. The contract requires timestamped forecasts, named decision owners, and three distinct recorded decisions. The probability history and closing date remain on the same page. Unlike the steward’s view, this regular account has no description-amendment form or Resolve control.",
    "verification": "Verified: contract expanded in the participant UI and absence of the Resolve button checked.",
    "image": "assets/screenshots/participant-workflow/C04-read-contract.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C6",
    "title": "Choose an outcome and inspect the fee",
    "caption": "Participant: click Trade, choose YES, enter 20 play credits, and open Terms. The screen exposes the initial trade fee of one credit and the probability-projection control. The highlighted purchase button names the selected outcome, making the order explicit before submission.",
    "verification": "Verified: these inputs were submitted in the next step. This screenshot precedes calculation of a projection; a computed later-order projection appears at the end of this appendix.",
    "image": "assets/screenshots/participant-workflow/C05-buy-preview.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C7",
    "title": "Confirm a purchase and see it in activity",
    "caption": "Participant: confirm the YES purchase. The accepted order appears in Bets with username, outcome, amount, resulting probability and timestamp. This first 20-credit YES purchase moved the displayed probability from 0.50 to about 0.83; available credit fell from 500 to 479 after the 20-credit order and one-credit initial fee.",
    "verification": "Verified: backend accepted the order and the live activity table displayed it. Values describe this synthetic trade, not a general pricing formula.",
    "image": "assets/screenshots/participant-workflow/C06-confirmed-trade.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C8",
    "title": "Distinguish ownership from executable selling",
    "caption": "Participant: revisit Trade and select Sell Shares. Immediately after the first purchase, the account owns shares but the UI reports that initial value cannot yet be sold without later activity from another participant. Displayed position value, share count and executable sale value are distinct; rounding can also prevent an otherwise plausible sale.",
    "verification": "Verified: first-trader guard reproduced. Additional synthetic participant orders were then submitted through the backend to establish the later sale scenario.",
    "image": "assets/screenshots/participant-workflow/C07-first-trader-sale-guard.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C9",
    "title": "Preview a permitted sale",
    "caption": "Participant: after further market activity, request a sale order of five credits and open Terms. Sale Preview shows the executable order, credits received, shares sold, value per share, dust assessment, and whether the order is allowed. The quote is informational; confirmation revalidates the current market state.",
    "verification": "Verified: quote allowed five shares to be sold for five credits, with zero dust in this scenario.",
    "image": "assets/screenshots/participant-workflow/C08-sale-preview.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C10",
    "title": "Confirm the sale and inspect positions",
    "caption": "Participant: confirm the sale and open Positions. The market’s position view refreshes alongside its probability and activity. The confirmed sale credited five play credits; remaining holdings are recalculated from market history, so share counts should not be inferred merely by subtracting the last order from an earlier screenshot.",
    "verification": "Verified: backend sale result reported sharesSold=5, netProceeds=5 and dust=0. Available credit increased from 479 to 484.",
    "image": "assets/screenshots/participant-workflow/C09-updated-position.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C11",
    "title": "Review holdings across markets",
    "caption": "Participant: open Profile, then Portfolio. This account shows one market and 33 YES shares after the demonstrated activity. The market link supports returning to the question, while summary counts and pagination make the portfolio usable across many questions. Financials is a separate account tab.",
    "verification": "Verified: portfolio readback matched the displayed 33 YES shares; no moderator market-management tabs are shown for this regular account.",
    "image": "assets/screenshots/participant-workflow/C10-participant-portfolio.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C12",
    "title": "Read a grouped question as a participant",
    "caption": "Participant: open the low-effort experiment group to compare its three independently traded options on one page. Each answer belongs to the same parent contract and topics, but retains its own YES/NO market. These are tests of several approaches, not mutually exclusive candidates for one winner.",
    "verification": "Verified: published group #2 displayed all three options from the moderator’s creation step. No grouped-answer trade was placed in this walkthrough.",
    "image": "assets/screenshots/participant-workflow/C11-grouped-user-view.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  },
  {
    "id": "C13",
    "title": "Calculate the effect of another order",
    "caption": "Participant: revisit the binary market, select another five-credit YES order, open Terms, and click Update Projection. The calculated projected probability is approximately 0.56 from a current value of approximately 0.53. This offers a view of market impact before deciding whether to trade again.",
    "verification": "Verified: projection response was currentProbability=0.5333333333 and projectedProbability=0.5625. This later preview was not submitted as another purchase.",
    "image": "assets/screenshots/participant-workflow/C12-next-trade-projection.png",
    "source": "https://docs.google.com/document/d/10b61aD8Mi0DX-5u4u8i-JXIVD8XKZX8fn7TsSKuxVkY/edit?tab=t.z1m4er8t1vu2"
  }
];
