SELECT t.id, t.name, t.description, t.up_votes, t.down_votes, (t.up_votes - t.down_votes >= $3) as confirmed
FROM translations t
WHERE t.word_id = $1
  AND t.language_id = (SELECT id FROM dic_languages WHERE lower(name) = lower($2))
  AND t.up_votes - t.down_votes >= $3
ORDER BY t.up_votes - t.down_votes DESC, t.id
LIMIT 1