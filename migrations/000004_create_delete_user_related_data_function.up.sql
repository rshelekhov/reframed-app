CREATE OR REPLACE FUNCTION delete_user_related_data(deleting_user_id varchar) RETURNS void AS $$
BEGIN
    DELETE FROM reminders WHERE user_id = deleting_user_id;
    DELETE FROM tags WHERE user_id = deleting_user_id;
    DELETE FROM tasks WHERE user_id = deleting_user_id;
    DELETE FROM headings WHERE user_id = deleting_user_id;
    DELETE FROM lists WHERE user_id = deleting_user_id;
END;
$$ LANGUAGE plpgsql;