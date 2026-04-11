CREATE TABLE IF NOT EXISTS tb_comment_diggs (
  id BIGINT NOT NULL AUTO_INCREMENT,
  uid BIGINT DEFAULT NULL,
  comment_id BIGINT DEFAULT NULL,
  create_time INT DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY ux_uid_comment (uid, comment_id),
  KEY idx_comment_uid (comment_id, uid),
  KEY idx_uid_create_time (uid, create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Deduplicate historical data if the table already existed without the unique key.
DELETE t1
FROM tb_comment_diggs t1
JOIN tb_comment_diggs t2
  ON t1.uid = t2.uid
 AND t1.comment_id = t2.comment_id
 AND t1.id > t2.id;

-- Rebuild comment digg counts from the relationship table.
UPDATE tb_comments c
LEFT JOIN (
  SELECT comment_id, COUNT(*) AS cnt
  FROM tb_comment_diggs
  GROUP BY comment_id
) d ON c.comment_id = d.comment_id
SET c.digg_count = COALESCE(d.cnt, 0);
