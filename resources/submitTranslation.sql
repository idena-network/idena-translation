SELECT ((t.val)::tp_submit_translation_result).res_code,
       ((t.val)::tp_submit_translation_result).translation_id
FROM (SELECT submit_translation($1, $2, $3, $4, $5, $6, $7) as val) t