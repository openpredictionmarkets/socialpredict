-- Example Markets for SocialPredict Platform
-- This file contains 10 diverse prediction markets that can be imported into PostgreSQL

-- Note: Assumes markets table structure matches the Go model:
-- id, created_at, updated_at, deleted_at, question_title, description, outcome_type, 
-- resolution_date_time, final_resolution_date_time, utc_offset, is_resolved, 
-- resolution_result, initial_probability, creator_username

INSERT INTO markets (
    question_title, 
    description, 
    outcome_type, 
    resolution_date_time, 
    final_resolution_date_time,
    utc_offset,
    is_resolved, 
    resolution_result, 
    initial_probability, 
    yes_label,
    no_label,
    creator_username,
    created_at,
    updated_at
) VALUES

-- 1. Technology Market
(
    'Will OpenAI release GPT-5 by end of 2025?',
    'This market resolves to YES if OpenAI officially announces and releases a model specifically named "GPT-5" or "GPT-5.0" by December 31, 2025, 11:59 PM UTC. The model must be publicly available or in limited beta testing. Early access programs count as release. Renamed versions (like GPT-4.5 Turbo Ultra) do not count unless explicitly called GPT-5.',
    'BINARY',
    '2025-12-31 23:59:59',
    '2026-01-05 23:59:59',
    0,
    false,
    null,
    0.35,
    'RELEASED',
    'NOT RELEASED',
    'admin',
    NOW(),
    NOW()
),

-- 2. Sports Market
(
    'Will Lionel Messi score 15+ goals in MLS 2025 regular season?',
    'This market resolves to YES if Lionel Messi scores 15 or more goals during the 2025 MLS regular season for Inter Miami CF. Only goals scored in official MLS regular season matches count. Playoff goals, friendly matches, international matches, and cup competitions do not count. Market resolves when the regular season ends or if Messi reaches 15 goals earlier.',
    'BINARY',
    '2025-11-15 23:59:59',
    '2025-11-20 23:59:59',
    -5,
    false,
    null,
    0.62,
    '15+ GOALS',
    'UNDER 15',
    'admin',
    NOW(),
    NOW()
),

-- 3. Politics Market
(
    'Will Donald Trump be the Republican nominee for President in 2028?',
    'This market resolves to YES if Donald Trump is officially nominated as the Republican Party candidate for President of the United States in the 2028 election. The nomination must be confirmed at the Republican National Convention. If Trump does not run, withdraws, or loses the nomination to another candidate, this resolves to NO.',
    'BINARY',
    '2028-08-31 23:59:59',
    '2028-09-05 23:59:59',
    -4,
    false,
    null,
    0.45,
    'TRUMP',
    'OTHER GOP',
    'admin',
    NOW(),
    NOW()
),

-- 4. Economics Market
(
    'Will Bitcoin price exceed $150,000 by end of 2025?',
    'This market resolves to YES if Bitcoin (BTC) price reaches or exceeds $150,000 USD on any major exchange (Coinbase, Binance, Kraken, or Bitstamp) at any point before December 31, 2025, 11:59 PM UTC. The price must be sustained for at least 1 hour on the exchange. Flash crashes or technical glitches lasting less than 1 hour do not count.',
    'BINARY',
    '2025-12-31 23:59:59',
    '2026-01-02 23:59:59',
    0,
    false,
    null,
    0.28,
    'BULL 🚀',
    'BEAR 📉',
    'admin',
    NOW(),
    NOW()
),

-- 5. Entertainment Market
(
    'Will Taylor Swift announce a new studio album in 2025?',
    'This market resolves to YES if Taylor Swift officially announces a new studio album (not re-recording, compilation, or live album) during 2025. The announcement must come from Taylor Swift herself, her official social media accounts, or her official representatives. The album does not need to be released in 2025, only announced.',
    'BINARY',
    '2025-12-31 23:59:59',
    '2026-01-05 23:59:59',
    0,
    false,
    null,
    0.72,
    'NEW ALBUM 🎵',
    'NO ALBUM',
    'admin',
    NOW(),
    NOW()
),

-- 6. Science Market
(
    'Will a cure for Type 1 Diabetes receive FDA approval by 2030?',
    'This market resolves to YES if the FDA approves a treatment that effectively cures Type 1 Diabetes (not just manages it) by December 31, 2030. The treatment must eliminate the need for insulin injections in at least 80% of patients for at least 2 years. Gene therapy, cell therapy, or artificial pancreas systems that meet these criteria count as cures.',
    'BINARY',
    '2030-12-31 23:59:59',
    '2031-01-07 23:59:59',
    0,
    false,
    null,
    0.15,
    'CURE FOUND',
    'NO CURE',
    'admin',
    NOW(),
    NOW()
),

-- 7. Weather/Climate Market
(
    'Will 2025 be the hottest year on record globally?',
    'This market resolves to YES if 2025 is declared the hottest year on record for global average temperature by NASA GISS, NOAA, or the UK Met Office. At least two of these three organizations must confirm 2025 as the hottest year. The announcement typically comes in January of the following year.',
    'BINARY',
    '2026-02-28 23:59:59',
    '2026-03-05 23:59:59',
    0,
    false,
    null,
    0.55,
    'HOTTEST 🔥',
    'COOLER',
    'admin',
    NOW(),
    NOW()
),

-- 8. Space Market
(
    'Will SpaceX successfully land humans on Mars by 2030?',
    'This market resolves to YES if SpaceX successfully lands at least one human being on the surface of Mars and that person survives the landing by December 31, 2030. The mission must be primarily operated by SpaceX, though partnerships with NASA or other organizations are allowed. The person must be alive for at least 24 hours after landing.',
    'BINARY',
    '2030-12-31 23:59:59',
    '2031-01-07 23:59:59',
    0,
    false,
    null,
    0.25,
    'MARS 🚀',
    'EARTH BOUND',
    'admin',
    NOW(),
    NOW()
),

-- 9. Business Market
(
    'Will Tesla stock price exceed $500 per share in 2025?',
    'This market resolves to YES if Tesla Inc. (TSLA) stock price reaches or exceeds $500.00 per share on NASDAQ during regular trading hours at any point in 2025. The price must be sustained for at least 15 minutes during market hours. After-hours trading and pre-market trading do not count. Stock splits will be adjusted accordingly.',
    'BINARY',
    '2025-12-31 23:59:59',
    '2026-01-02 23:59:59',
    -5,
    false,
    null,
    0.38,
    'MOON 📈',
    'GROUNDED',
    'admin',
    NOW(),
    NOW()
),

-- 10. Gaming Market
(
    'Will Grand Theft Auto 6 be released in 2025?',
    'This market resolves to YES if Rockstar Games releases Grand Theft Auto 6 (GTA 6) for any gaming platform in 2025. The game must be available for purchase by consumers, not just announced or in beta. Early access programs count as release. The game must be specifically titled "Grand Theft Auto 6" or "Grand Theft Auto VI".',
    'BINARY',
    '2025-12-31 23:59:59',
    '2026-01-05 23:59:59',
    0,
    false,
    null,
    0.42,
    'RELEASED 🎮',
    'DELAYED',
    'admin',
    NOW(),
    NOW()
);