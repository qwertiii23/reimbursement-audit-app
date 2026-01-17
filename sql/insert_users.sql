INSERT INTO `users` (`id`, `username`, `password`, `email`, `real_name`, `role`, `status`, `created_at`, `updated_at`, `last_login`) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'admin', '$2a$10$Kq5rqflBxeycDvpGVmsDHeVmnk/zV8mTLRVvVnq.tCjAG3XOpq7By', 'admin@example.com', '系统管理员', 'admin', 'active', NOW(), NOW(), NULL),
('550e8400-e29b-41d4-a716-446655440002', 'user', '$2a$10$1KwZZWwsaHvk2VEszFGSDeJB1WsSthAQ0s00yclXkVb9pGbS9FVKq', 'user@example.com', '普通用户', 'user', 'active', NOW(), NOW(), NULL),
('550e8400-e29b-41d4-a716-446655440003', 'test', '$2a$10$BPBRMxokIyKdoqP1v4vcxOMdP3QHCNCMk/qHc.DPuNZM/2a2xZrOy', 'test@example.com', '测试用户', 'user', 'active', NOW(), NOW(), NULL);
