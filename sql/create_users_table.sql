CREATE TABLE IF NOT EXISTS `users` (
  `id` varchar(36) NOT NULL COMMENT '用户ID',
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '密码（加密）',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `real_name` varchar(50) DEFAULT NULL COMMENT '真实姓名',
  `role` varchar(20) NOT NULL DEFAULT 'user' COMMENT '角色（admin/user）',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态（active/inactive/locked）',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `last_login` datetime DEFAULT NULL COMMENT '最后登录时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  UNIQUE KEY `idx_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
