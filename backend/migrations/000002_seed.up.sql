-- Password is 'password123' (bcrypt cost 12)
INSERT INTO users (id, name, email, password) VALUES 
('11111111-1111-1111-1111-111111111111', 'Admin User', 'test@example.com', '$2a$12$HOIDJETAzbfMQz20rHgiAu9FBVgTezE18cw9G/b6ZCryxiYJl2lYW');

INSERT INTO projects (id, name, description, owner_id) VALUES 
('22222222-2222-2222-2222-222222222222', 'Alpha Core', 'Main platform development', '11111111-1111-1111-1111-111111111111');

INSERT INTO tasks (title, description, status, priority, project_id, creator_id, assignee_id) VALUES 
('Setup Database', 'Write initial SQL migrations', 'done', 'high', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111'),
('Implement Auth', 'JWT based authentication', 'in_progress', 'high', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111'),
('Design API', 'Create RESTful endpoints', 'todo', 'medium', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111');