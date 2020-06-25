SELECT ((t.val)::tp_vote_result).res_code,
       ((t.val)::tp_vote_result).up_votes,
       ((t.val)::tp_vote_result).down_votes
FROM (SELECT vote($1, $2, $3, $4) as val) t