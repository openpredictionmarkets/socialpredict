-- Example Sports Markets for SocialPredict
-- 30 diverse sports betting markets covering various sports and scenarios
-- All markets are for future events and created by admin user
-- Follows the Market model schema with proper column names

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

-- NFL Markets (2025-2026 Season)
('Will the Kansas City Chiefs win Super Bowl LX?', 'Market resolves YES if the Kansas City Chiefs win Super Bowl LX in February 2026. Resolves NO if any other team wins.', 'BINARY', '2026-02-08 23:59:00', '2026-02-12 23:59:00', 0, false, '', 0.32, 'CHIEFS WIN 🏆', 'CHIEFS LOSE ❌', 'admin', NOW(), NOW()),

('Will Josh Allen throw for over 4,000 yards in 2025 NFL season?', 'Market resolves YES if Josh Allen throws for more than 4,000 passing yards in the 2025 NFL regular season. Playoff stats do not count.', 'BINARY', '2026-01-05 23:59:00', '2026-01-08 23:59:00', 0, false, '', 0.68, 'OVER 4,000 📈', 'UNDER 4,000 📉', 'admin', NOW(), NOW()),

('Will there be a tie game in the 2025 NFL season?', 'Market resolves YES if any NFL regular season game ends in a tie during the 2025 season. Overtime rules must result in a tied final score.', 'BINARY', '2026-01-05 23:59:00', '2026-01-08 23:59:00', 0, false, '', 0.25, 'TIE GAME 🤝', 'NO TIES ❌', 'admin', NOW(), NOW()),

-- NBA Markets (2025-2026 Season)
('Will Victor Wembanyama win MVP in 2025-26?', 'Market resolves YES if Victor Wembanyama wins the NBA Most Valuable Player award for the 2025-26 regular season.', 'BINARY', '2026-05-15 23:59:00', '2026-05-20 23:59:00', 0, false, '', 0.18, 'WEMBY MVP 🏀', 'NO MVP ❌', 'admin', NOW(), NOW()),

('Will the Denver Nuggets win the 2026 NBA Championship?', 'Market resolves YES if the Denver Nuggets win the 2026 NBA Championship. Resolves NO if any other team wins the title.', 'BINARY', '2026-06-20 23:59:00', '2026-06-25 23:59:00', 0, false, '', 0.22, 'NUGGETS CHAMP 🏆', 'NO NUGGETS ❌', 'admin', NOW(), NOW()),

('Will Luka Doncic average 30+ PPG in 2025-26 season?', 'Market resolves YES if Luka Doncic averages 30.0 or more points per game during the 2025-26 NBA regular season.', 'BINARY', '2026-04-15 23:59:00', '2026-04-20 23:59:00', 0, false, '', 0.45, 'LUKA 30+ 🔥', 'UNDER 30 📉', 'admin', NOW(), NOW()),

-- MLB Markets (2026 Season)
('Will Aaron Judge hit 70+ home runs in 2026?', 'Market resolves YES if Aaron Judge hits 70 or more home runs during the 2026 MLB regular season. Spring training and playoff homers do not count.', 'BINARY', '2026-09-30 23:59:00', '2026-10-05 23:59:00', 0, false, '', 0.08, '70+ BOMBS 💥', 'UNDER 70 📉', 'admin', NOW(), NOW()),

('Will the New York Yankees win the 2026 World Series?', 'Market resolves YES if the New York Yankees win the 2026 World Series. Resolves NO if any other team wins.', 'BINARY', '2026-11-01 23:59:00', '2026-11-05 23:59:00', 0, false, '', 0.18, 'YANKEES WIN ⚾', 'NO YANKEES ❌', 'admin', NOW(), NOW()),

('Will there be a no-hitter in the 2026 MLB season?', 'Market resolves YES if any pitcher throws a no-hitter (no hits allowed) during the 2026 MLB regular season or playoffs.', 'BINARY', '2026-11-01 23:59:00', '2026-11-05 23:59:00', 0, false, '', 0.78, 'NO-HITTER ⭐', 'NO NO-HITTER ❌', 'admin', NOW(), NOW()),

-- Soccer/Football Markets (2025-2026)
('Will Real Madrid win the 2026 Champions League?', 'Market resolves YES if Real Madrid wins the UEFA Champions League 2025-26 tournament. Resolves NO if any other team wins.', 'BINARY', '2026-05-31 23:59:00', '2026-06-05 23:59:00', 0, false, '', 0.25, 'REAL MADRID 👑', 'NOT REAL ❌', 'admin', NOW(), NOW()),

('Will England win Euro 2028?', 'Market resolves YES if England wins the UEFA European Championship 2028. Resolves NO if any other team wins.', 'BINARY', '2028-07-31 23:59:00', '2028-08-05 23:59:00', 0, false, '', 0.35, 'ENGLAND WINS ⚽', 'NOT ENGLAND ❌', 'admin', NOW(), NOW()),

('Will Kylian Mbappe score 40+ goals in 2025-26 season?', 'Market resolves YES if Kylian Mbappe scores 40 or more goals in all competitions during the 2025-26 season.', 'BINARY', '2026-05-31 23:59:00', '2026-06-03 23:59:00', 0, false, '', 0.55, 'MBAPPE 40+ ⚽', 'UNDER 40 📉', 'admin', NOW(), NOW()),

-- Tennis Markets (2026)
('Will Carlos Alcaraz win Wimbledon 2026?', 'Market resolves YES if Carlos Alcaraz wins the Wimbledon Championships 2026 mens singles title. Resolves NO if any other player wins.', 'BINARY', '2026-07-15 23:59:00', '2026-07-18 23:59:00', 0, false, '', 0.32, 'ALCARAZ WINS 🎾', 'NOT ALCARAZ ❌', 'admin', NOW(), NOW()),

('Will Iga Swiatek win the 2026 French Open?', 'Market resolves YES if Iga Swiatek wins the 2026 French Open womens singles title. Resolves NO if any other player wins.', 'BINARY', '2026-06-08 23:59:00', '2026-06-12 23:59:00', 0, false, '', 0.48, 'SWIATEK WINS 🏆', 'NOT SWIATEK ❌', 'admin', NOW(), NOW()),

-- Hockey Markets (2025-2026)
('Will the Florida Panthers repeat as Stanley Cup Champions?', 'Market resolves YES if the Florida Panthers win the 2026 Stanley Cup Championship. Resolves NO if any other team wins.', 'BINARY', '2026-06-30 23:59:00', '2026-07-05 23:59:00', 0, false, '', 0.15, 'PANTHERS REPEAT 🏒', 'NO REPEAT ❌', 'admin', NOW(), NOW()),

('Will Connor Bedard win the Hart Trophy in 2025-26?', 'Market resolves YES if Connor Bedard wins the Hart Memorial Trophy (NHL MVP) for the 2025-26 season.', 'BINARY', '2026-06-15 23:59:00', '2026-06-20 23:59:00', 0, false, '', 0.12, 'BEDARD MVP 👑', 'NO MVP ❌', 'admin', NOW(), NOW()),

-- Golf Markets (2026)
('Will Tiger Woods win a major championship in 2026?', 'Market resolves YES if Tiger Woods wins any of the four major championships (Masters, PGA, US Open, British Open) in 2026.', 'BINARY', '2026-07-31 23:59:00', '2026-08-05 23:59:00', 0, false, '', 0.03, 'TIGER MAJOR 🐅', 'NO MAJOR ❌', 'admin', NOW(), NOW()),

('Will there be a playoff at The Masters 2026?', 'Market resolves YES if The Masters Tournament 2026 ends in a playoff between multiple players. Resolves NO if there is a clear winner after 72 holes.', 'BINARY', '2026-04-12 23:59:00', '2026-04-15 23:59:00', 0, false, '', 0.28, 'PLAYOFF 🏌️', 'NO PLAYOFF ❌', 'admin', NOW(), NOW()),

-- Formula 1 Markets (2026)
('Will Ferrari win the 2026 F1 Constructors Championship?', 'Market resolves YES if Ferrari wins the 2026 Formula 1 Constructors Championship. Resolves NO if any other team wins.', 'BINARY', '2026-12-08 23:59:00', '2026-12-12 23:59:00', 0, false, '', 0.28, 'FERRARI WINS 🏎️', 'NOT FERRARI ❌', 'admin', NOW(), NOW()),

('Will there be a new F1 race winner in 2026?', 'Market resolves YES if a driver wins their first Formula 1 race during the 2026 season. Resolves NO if only previous race winners win in 2026.', 'BINARY', '2026-12-08 23:59:00', '2026-12-12 23:59:00', 0, false, '', 0.65, 'NEW WINNER 🏆', 'NO NEW WINNER ❌', 'admin', NOW(), NOW()),

-- Boxing/MMA Markets (2026)
('Will Ryan Garcia fight Gervonta Davis in 2026?', 'Market resolves YES if Ryan Garcia and Gervonta Davis have a professional boxing match against each other during 2026.', 'BINARY', '2026-12-31 23:59:00', '2027-01-05 23:59:00', 0, false, '', 0.42, 'FIGHT HAPPENS 🥊', 'NO FIGHT ❌', 'admin', NOW(), NOW()),

('Will Alex Pereira defend his UFC title 3+ times in 2026?', 'Market resolves YES if Alex Pereira successfully defends his UFC Light Heavyweight Championship 3 or more times during 2026.', 'BINARY', '2026-12-31 23:59:00', '2027-01-05 23:59:00', 0, false, '', 0.35, '3+ DEFENSES 👑', 'UNDER 3 ❌', 'admin', NOW(), NOW()),

-- College Sports Markets
('Will UConn win the 2026 NCAA Basketball Tournament?', 'Market resolves YES if the University of Connecticut mens basketball team wins the 2026 NCAA Division I Basketball Tournament.', 'BINARY', '2026-04-08 23:59:00', '2026-04-12 23:59:00', 0, false, '', 0.08, 'UCONN WINS 🏀', 'NOT UCONN ❌', 'admin', NOW(), NOW()),

('Will LSU win the 2026 College Football National Championship?', 'Market resolves YES if Louisiana State University wins the 2026 College Football National Championship. Resolves NO if any other team wins.', 'BINARY', '2027-01-20 23:59:00', '2027-01-25 23:59:00', 0, false, '', 0.12, 'LSU CHAMPS 🐅', 'NOT LSU ❌', 'admin', NOW(), NOW()),

-- Esports Markets (2026)
('Will League of Legends Worlds 2026 have 100M+ viewers?', 'Market resolves YES if the League of Legends World Championship 2026 achieves 100 million or more peak concurrent viewers across all platforms.', 'BINARY', '2026-11-15 23:59:00', '2026-11-20 23:59:00', 0, false, '', 0.72, '100M+ VIEWERS 📺', 'UNDER 100M ❌', 'admin', NOW(), NOW()),

('Will a Western team win Worlds 2026 in League of Legends?', 'Market resolves YES if a team from Europe or North America wins the League of Legends World Championship 2026.', 'BINARY', '2026-11-15 23:59:00', '2026-11-20 23:59:00', 0, false, '', 0.25, 'WEST WINS 🌍', 'EAST WINS 🌏', 'admin', NOW(), NOW()),

-- Swimming/Track Markets
('Will Katie Ledecky win gold at 2028 Olympics?', 'Market resolves YES if Katie Ledecky wins at least one gold medal in swimming at the 2028 Los Angeles Olympics.', 'BINARY', '2028-08-12 23:59:00', '2028-08-17 23:59:00', 0, false, '', 0.85, 'LEDECKY GOLD 🥇', 'NO GOLD ❌', 'admin', NOW(), NOW()),

('Will the marathon world record be broken in 2026?', 'Market resolves YES if either the mens or womens marathon world record is officially broken during 2026 in any sanctioned race.', 'BINARY', '2026-12-31 23:59:00', '2027-01-05 23:59:00', 0, false, '', 0.22, 'RECORD BROKEN 🏃', 'NO RECORD ❌', 'admin', NOW(), NOW()),

-- Winter Sports Markets
('Will Norway top the medal table at 2026 Winter Olympics?', 'Market resolves YES if Norway wins the most total medals at the 2026 Winter Olympics in Milan-Cortina. Resolves NO if any other country wins more medals.', 'BINARY', '2026-02-28 23:59:00', '2026-03-05 23:59:00', 0, false, '', 0.58, 'NORWAY TOPS 🇳🇴', 'NOT NORWAY ❌', 'admin', NOW(), NOW()),

('Will Mikaela Shiffrin win 100+ World Cup races by end of 2026?', 'Market resolves YES if Mikaela Shiffrin reaches 100 or more Alpine Ski World Cup race victories by December 31, 2026.', 'BINARY', '2026-12-31 23:59:00', '2027-01-05 23:59:00', 0, false, '', 0.78, '100+ WINS 🎿', 'UNDER 100 ❌', 'admin', NOW(), NOW()),

-- Unique Achievement Markets
('Will any team go undefeated in 2026 NFL regular season?', 'Market resolves YES if any NFL team finishes the 2026 regular season with a perfect 17-0 record. Playoff games do not count.', 'BINARY', '2027-01-10 23:59:00', '2027-01-15 23:59:00', 0, false, '', 0.02, 'PERFECT SEASON 💯', 'NO PERFECT ❌', 'admin', NOW(), NOW()),

('Will a pitcher throw an immaculate inning in 2026 MLB?', 'Market resolves YES if any pitcher throws an immaculate inning (9 strikes, 3 strikeouts) during the 2026 MLB regular season or playoffs.', 'BINARY', '2026-11-01 23:59:00', '2026-11-05 23:59:00', 0, false, '', 0.35, 'IMMACULATE ⚾', 'NO IMMACULATE ❌', 'admin', NOW(), NOW());