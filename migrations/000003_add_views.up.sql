CREATE VIEW task_tags_view AS
SELECT task_id, ARRAY_AGG(DISTINCT title) AS tags
FROM tasks_tags
         JOIN tags ON tasks_tags.tag_id = tags.id
GROUP BY task_id;