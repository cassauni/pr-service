-- должны быть созданы до вставки
-- CREATE TABLE teams (
--   team_name text PRIMARY KEY
-- );
--
-- CREATE TABLE users (
--   user_id   text PRIMARY KEY,
--   username  text NOT NULL,
--   team_name text NOT NULL REFERENCES teams(team_name),
--   is_active boolean NOT NULL DEFAULT true
-- );
-- === SEED: команды ===
INSERT INTO teams (team_name) VALUES
                                  ('backend'),
                                  ('frontend'),
                                  ('mobile_ios'),
                                  ('mobile_android'),
                                  ('devops'),
                                  ('qa'),
                                  ('payments'),
                                  ('billing'),
                                  ('risk'),
                                  ('analytics'),
                                  ('data_platform'),
                                  ('ml'),
                                  ('infra'),
                                  ('core'),
                                  ('search'),
                                  ('recommendations'),
                                  ('content'),
                                  ('internal_tools'),
                                  ('support'),
                                  ('marketing')
    ON CONFLICT (team_name) DO NOTHING;
-- === SEED: пользователи (200 шт, до 10 на команду) ===
WITH numbered_teams AS (
    SELECT
        team_name,
        row_number() OVER (ORDER BY team_name) AS team_idx
    FROM teams
    WHERE team_name IN (
                        'backend',
                        'frontend',
                        'mobile_ios',
                        'mobile_android',
                        'devops',
                        'qa',
                        'payments',
                        'billing',
                        'risk',
                        'analytics',
                        'data_platform',
                        'ml',
                        'infra',
                        'core',
                        'search',
                        'recommendations',
                        'content',
                        'internal_tools',
                        'support',
                        'marketing'
        )
),
     nums AS (
         -- для каждой команды генерируем 10 пользователей
         SELECT
             t.team_name,
             (t.team_idx - 1) * 10 + gs AS n
         FROM numbered_teams t
                  CROSS JOIN generate_series(1, 10) AS gs
     )
INSERT INTO users (user_id, username, team_name, is_active)
SELECT
    format('u%03s', n)                             AS user_id,
    format('User %03s', n)                         AS username,
    team_name,
    CASE WHEN n % 5 = 0 THEN FALSE ELSE TRUE END   AS is_active
FROM nums
ORDER BY n;
