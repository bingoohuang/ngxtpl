DROP TABLE IF EXISTS `t_upstreams`;

CREATE TABLE `t_upstreams` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `name` varchar(60) NOT NULL COMMENT 'upstream名称',
  `keepalive` varchar(60)  NULL COMMENT '连接池',
  `ip_hash` varchar(60)  NULL COMMENT 'ip_hash yes or other',
  `resolver` varchar(60)  NULL COMMENT 'resolver',
  `state` char(1)  DEFAULT '1' COMMENT '状态 1:有效 0:无效',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `t_servers`;

CREATE TABLE `t_servers` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID。',
  `upstream_name` varchar(60) NOT NULL COMMENT 'upstream名称',
  `address` varchar(60) NOT NULL COMMENT '服务器地址。',
  `port` varchar(60) NOT NULL DEFAULT '80' COMMENT '端口号。',
  `weight` varchar(60)  NULL  COMMENT '权重。',
  `max_conns` varchar(60)  NULL   COMMENT '并发最大连接数。',
  `max_fails` varchar(60)  NULL  COMMENT '并发失败尝试次数。',
  `fail_timeout` varchar(20) NOT NULL DEFAULT '10s' COMMENT '失败超时。',
  `backup` varchar(60)  NULL COMMENT '是否备份服务。 yes or other',
  `down` varchar(60)  NULL  COMMENT '状态: yes or other',
  `slow_start` varchar(20) NULL COMMENT '慢启动时间。',
  `state` char(1)  DEFAULT '1' COMMENT '状态 1:有效 0:无效',
  PRIMARY KEY (`id`),
  KEY `upstream_name` (`upstream_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

insert into t_upstreams(`id`, `name`, `keepalive`, ip_hash, `resolver`, `state`) values
('1', 'service1', '32', '', '', '1');

insert into t_servers(`id`, upstream_name, `address`, `port`, `weight`, max_conns, max_fails, fail_timeout, `backup`, `down`, slow_start, `state`) values
('1', 'service1', '127.0.0.1', '1001', '', '', '', '', '', '', '', '1'),
('2', 'service1', '127.0.0.1', '1002', '', '', '', '', '', '', '', '1');