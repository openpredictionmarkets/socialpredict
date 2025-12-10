// AUTO-GENERATED TYPES
// This file was created by bin/scaffoldSDK.ts based on backend/docs/openapi.yaml.
//
// It mirrors the shapes under components/schemas as TypeScript interfaces
// and types for use by the generated SDK stubs.

export interface ErrorResponse {
  error: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  username: string;
  usertype: string;
  mustChangePassword: boolean;
}

export interface CreateMarketRequest {
  questionTitle: string;
  description?: string;
  outcomeType: string;
  resolutionDateTime: string;
  yesLabel?: string;
  noLabel?: string;
}

export interface MarketResponse {
  id?: number;
  questionTitle?: string;
  description?: string;
  outcomeType?: string;
  resolutionDateTime?: string;
  creatorUsername?: string;
  yesLabel?: string;
  noLabel?: string;
  status?: string;
  isResolved?: boolean;
  resolutionResult?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreatorResponse {
  username?: string;
  personalEmoji?: string;
  displayname?: string;
}

export interface MarketOverviewResponse {
  market?: MarketResponse;
  creator?: CreatorResponse;
  lastProbability?: number;
  numUsers?: number;
  totalVolume?: number;
  marketDust?: number;
}

export interface ListMarketsResponse {
  markets?: MarketOverviewResponse[];
  total?: number;
}

export interface MarketDetailsResponse {
  market?: PublicMarketResponse;
  creator?: CreatorResponse;
  probabilityChanges?: ProbabilityChange[];
  numUsers?: number;
  totalVolume?: number;
  marketDust?: number;
}

export interface PublicMarketResponse {
  id?: number;
  questionTitle?: string;
  description?: string;
  outcomeType?: string;
  resolutionDateTime?: string;
  finalResolutionDateTime?: string;
  utcOffset?: number;
  isResolved?: boolean;
  resolutionResult?: string;
  initialProbability?: number;
  creatorUsername?: string;
  createdAt?: string;
  yesLabel?: string;
  noLabel?: string;
}

export interface ProbabilityChange {
  probability?: number;
  timestamp?: string;
}

export interface SearchResponse {
  primaryResults?: MarketOverviewResponse[];
  fallbackResults?: MarketOverviewResponse[];
  query?: string;
  primaryStatus?: string;
  primaryCount?: number;
  fallbackCount?: number;
  totalCount?: number;
  fallbackUsed?: boolean;
}

export interface ResolveMarketRequest {
  resolution: "yes" | "no";
}

export interface MarketLeaderboardResponse {
  marketId?: number;
  leaderboard?: LeaderboardRow[];
  total?: number;
}

export interface LeaderboardRow {
  username?: string;
  profit?: number;
  currentValue?: number;
  totalSpent?: number;
  position?: string;
  yesSharesOwned?: number;
  noSharesOwned?: number;
  rank?: number;
}

export interface ProbabilityProjectionResponse {
  marketId?: number;
  currentProbability?: number;
  projectedProbability?: number;
  amount?: number;
  outcome?: string;
}

export interface PlaceBetRequest {
  marketId?: number;
  amount?: number;
  outcome?: string;
}

export interface PlaceBetResponse {
  username?: string;
  marketId?: number;
  amount?: number;
  outcome?: string;
  placedAt?: string;
}

export interface SellBetRequest {
  marketId?: number;
  amount?: number;
  outcome?: string;
}

export interface SellBetResponse {
  username?: string;
  marketId?: number;
  sharesSold?: number;
  saleValue?: number;
  dust?: number;
  outcome?: string;
  transactionAt?: string;
}

export interface MarketPosition {
  username?: string;
  yesSharesOwned?: number;
  noSharesOwned?: number;
  value?: number;
}

export interface UserPosition {
  username?: string;
  marketId?: number;
  yesSharesOwned?: number;
  noSharesOwned?: number;
  value?: number;
  totalSpent?: number;
  totalSpentInPlay?: number;
  isResolved?: boolean;
  resolutionResult?: string;
}

export interface MarketBet {
  username?: string;
  marketId?: number;
  amount?: number;
  outcome?: string;
  placedAt?: string;
}

export interface PrivateUserResponse {
  id?: number;
  username?: string;
  displayname?: string;
  usertype?: string;
  initialAccountBalance?: number;
  accountBalance?: number;
  personalEmoji?: string;
  description?: string;
  personalink1?: string;
  personalink2?: string;
  personalink3?: string;
  personalink4?: string;
  email?: string;
  apiKey?: string;
  mustChangePassword?: boolean;
}

export interface PublicUserResponse {
  username?: string;
  displayname?: string;
  usertype?: string;
  initialAccountBalance?: number;
  accountBalance?: number;
  personalEmoji?: string;
  description?: string;
  personalink1?: string;
  personalink2?: string;
  personalink3?: string;
  personalink4?: string;
}

export interface UserCreditResponse {
  credit?: number;
}

export interface PortfolioItem {
  marketId?: number;
  questionTitle?: string;
  yesSharesOwned?: number;
  noSharesOwned?: number;
  lastBetPlaced?: string;
}

export interface PortfolioResponse {
  portfolioItems?: PortfolioItem[];
  totalSharesOwned?: number;
}

export interface AdminCreateUserRequest {
  username?: string;
}

export interface AdminCreateUserResponse {
  message?: string;
  username?: string;
  password?: string;
  usertype?: string;
}

export interface UserFinancialResponse {
  financial?: { [key: string]: number };
}

export interface ChangeDescriptionRequest {
  description?: string;
}

export interface ChangeDisplayNameRequest {
  displayName?: string;
}

export interface ChangeEmojiRequest {
  emoji?: string;
}

export interface ChangePersonalLinksRequest {
  personalLink1?: string;
  personalLink2?: string;
  personalLink3?: string;
  personalLink4?: string;
}

export interface ChangePasswordRequest {
  currentPassword?: string;
  newPassword?: string;
}

export interface Economics {
  marketcreation?: MarketCreation;
  marketincentives?: MarketIncentives;
  user?: UserEconomics;
  betting?: Betting;
}

export interface MarketCreation {
  initialMarketProbability?: number;
  initialMarketSubsidization?: number;
  initialMarketYes?: number;
  initialMarketNo?: number;
  minimumFutureHours?: number;
}

export interface MarketIncentives {
  createMarketCost?: number;
  traderBonus?: number;
}

export interface UserEconomics {
  initialAccountBalance?: number;
  maximumDebtAllowed?: number;
}

export interface BetFees {
  initialBetFee?: number;
  buySharesFee?: number;
  sellSharesFee?: number;
}

export interface Betting {
  minimumBet?: number;
  maxDustPerSale?: number;
  betFees?: BetFees;
}

export interface FrontendConfig {
  charts?: FrontendCharts;
}

export interface FrontendCharts {
  sigFigs?: number;
}

export interface GlobalLeaderboardEntry {
  username?: string;
  totalProfit?: number;
  totalCurrentValue?: number;
  totalSpent?: number;
  activeMarkets?: number;
  resolvedMarkets?: number;
  earliestBet?: string;
  rank?: number;
}

export interface HomeContent {
  title?: string;
  format?: string;
  html?: string;
  markdown?: string;
  version?: number;
  updatedAt?: string;
}

export interface HomeContentUpdateRequest {
  title?: string;
  format?: string;
  markdown?: string;
  html?: string;
  version?: number;
}
