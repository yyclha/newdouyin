-- Active: 1765185998682@@127.0.0.1@3306@db_douyin
-- MySQL dump
-- Host: localhost    Database: db_douyin

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for tb_accounts
-- ----------------------------
DROP TABLE IF EXISTS `tb_accounts`;
CREATE TABLE `tb_accounts` (
  `uid` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `nickname` varchar(100) DEFAULT NULL COMMENT '昵称',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `password` varchar(128) DEFAULT NULL COMMENT '密码',
  `last_login_ip` varchar(100) DEFAULT NULL COMMENT '最后登录IP',
  `create_time` bigint(20) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户表';

-- ----------------------------
-- Table structure for tb_users
-- ----------------------------
DROP TABLE IF EXISTS `tb_users`;
CREATE TABLE `tb_users` (
  `uid` bigint(20) NOT NULL COMMENT '用户ID',
  `short_id` int(11) DEFAULT NULL COMMENT '短ID',
  `unique_id` varchar(255) DEFAULT NULL COMMENT '唯一ID',
  `gender` char(1) DEFAULT NULL COMMENT '性别',
  `user_age` int(11) DEFAULT NULL COMMENT '年龄',
  `nickname` varchar(100) DEFAULT NULL COMMENT '昵称',
  `country` varchar(100) DEFAULT NULL COMMENT '国家',
  `province` varchar(100) DEFAULT NULL COMMENT '省份',
  `district` varchar(255) DEFAULT NULL COMMENT '地区',
  `city` varchar(255) DEFAULT NULL COMMENT '城市',
  `signature` text COMMENT '签名',
  `ip_location` varchar(100) DEFAULT NULL COMMENT 'IP归属地',
  `birthday_hide_level` int(11) DEFAULT NULL COMMENT '生日隐藏等级',
  `can_show_group_card` int(11) DEFAULT NULL COMMENT '是否显示群名片',
  `aweme_count` bigint(20) DEFAULT NULL COMMENT '作品数量',
  `total_favorited` bigint(20) DEFAULT NULL COMMENT '总获赞数',
  `favoriting_count` int(11) DEFAULT NULL COMMENT '喜欢数',
  `follower_count` bigint(20) DEFAULT NULL COMMENT '粉丝数',
  `following_count` bigint(20) DEFAULT NULL COMMENT '关注数',
  `forward_count` int(11) DEFAULT NULL COMMENT '转发数',
  `public_collects_count` int(11) DEFAULT NULL COMMENT '公开收藏数',
  `mplatform_followers_count` bigint(20) DEFAULT NULL COMMENT '全平台粉丝数',
  `max_follower_count` bigint(20) DEFAULT NULL COMMENT '最大粉丝数',
  `follow_status` int(11) DEFAULT NULL COMMENT '关注状态',
  `follower_status` int(11) DEFAULT NULL COMMENT '粉丝状态',
  `follower_request_status` int(11) DEFAULT NULL COMMENT '粉丝请求状态',
  `cover_colour` varchar(100) DEFAULT NULL COMMENT '封面颜色',
  `cover_url` json DEFAULT NULL COMMENT '封面URL',
  `white_cover_url` json DEFAULT NULL COMMENT '白色封面URL',
  `share_info` json DEFAULT NULL COMMENT '分享信息',
  `commerce_info` json DEFAULT NULL COMMENT '商业信息',
  `commerce_user_info` json DEFAULT NULL COMMENT '商业用户信息',
  `commerce_user_level` int(11) DEFAULT NULL COMMENT '商业用户等级',
  `card_entries` json DEFAULT NULL COMMENT '卡片入口',
  `avatar_168x168` json DEFAULT NULL COMMENT '168x168头像',
  `avatar_300x300` json DEFAULT NULL COMMENT '300x300头像',
  `avatar_small` json DEFAULT NULL COMMENT '小头像',
  `avatar_large` json DEFAULT NULL COMMENT '大头像',
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';

-- ----------------------------
-- Table structure for tb_relations
-- ----------------------------
DROP TABLE IF EXISTS `tb_relations`;
CREATE TABLE `tb_relations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '关系ID',
  `follower_id` bigint(20) DEFAULT NULL COMMENT '粉丝ID',
  `following_id` bigint(20) DEFAULT NULL COMMENT '被关注者ID',
  `create_time` bigint(20) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_relation` (`follower_id`, `following_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户关系表';

-- ----------------------------
-- Table structure for tb_messages
-- ----------------------------
DROP TABLE IF EXISTS `tb_messages`;
CREATE TABLE `tb_messages` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `tx_uid` bigint(20) NOT NULL COMMENT '发送者ID',
  `rx_uid` bigint(20) NOT NULL COMMENT '接收者ID',
  `msg_type` int(11) DEFAULT NULL COMMENT '消息类型',
  `msg_data` text COMMENT '消息内容',
  `read_state` int(11) DEFAULT NULL COMMENT '读取状态',
  `create_time` int(11) DEFAULT NULL COMMENT '创建时间',
  `delete_time` int(11) DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';

-- ----------------------------
-- Table structure for tb_posts
-- ----------------------------
DROP TABLE IF EXISTS `tb_posts`;
CREATE TABLE `tb_posts` (
  `id` varchar(100) NOT NULL COMMENT '帖子ID',
  `model_type` varchar(100) DEFAULT NULL COMMENT '模型类型',
  `note_card` json DEFAULT NULL COMMENT '笔记卡片',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='帖子表';

-- ----------------------------
-- Table structure for tb_goods
-- ----------------------------
DROP TABLE IF EXISTS `tb_goods`;
CREATE TABLE `tb_goods` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '商品ID',
  `name` varchar(255) DEFAULT NULL COMMENT '商品名称',
  `cover` varchar(255) DEFAULT NULL COMMENT '封面',
  `imgs` text COMMENT '图片列表',
  `is_low_price` tinyint(1) DEFAULT NULL COMMENT '是否低价',
  `discount` varchar(100) DEFAULT NULL COMMENT '折扣',
  `sold` double DEFAULT NULL COMMENT '已售数量',
  `price` double DEFAULT NULL COMMENT '价格',
  `real_price` double DEFAULT NULL COMMENT '真实价格',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';

-- ----------------------------
-- Table structure for tb_videos
-- ----------------------------
DROP TABLE IF EXISTS `tb_videos`;
CREATE TABLE `tb_videos` (
  `aweme_id` bigint(20) NOT NULL COMMENT '视频ID',
  `video_desc` text COMMENT '视频描述',
  `create_time` int(11) DEFAULT NULL COMMENT '创建时间',
  `share_url` varchar(255) DEFAULT NULL COMMENT '分享链接',
  `status` json DEFAULT NULL COMMENT '状态',
  `text_extra` json DEFAULT NULL COMMENT '文本扩展信息',
  `is_top` int(11) DEFAULT NULL COMMENT '是否置顶',
  `share_info` json DEFAULT NULL COMMENT '分享信息',
  `duration` int(11) DEFAULT NULL COMMENT '时长',
  `image_infos` json DEFAULT NULL COMMENT '图片信息',
  `risk_infos` json DEFAULT NULL COMMENT '风险信息',
  `position` json DEFAULT NULL COMMENT '位置信息',
  `author_user_id` bigint(20) DEFAULT NULL COMMENT '作者ID',
  `prevent_download` tinyint(4) DEFAULT NULL COMMENT '是否禁止下载',
  `long_video` json DEFAULT NULL COMMENT '长视频信息',
  `aweme_control` json DEFAULT NULL COMMENT '视频控制信息',
  `images` json DEFAULT NULL COMMENT '图片列表',
  `suggest_words` json DEFAULT NULL COMMENT '建议词',
  `video_tag` json DEFAULT NULL COMMENT '视频标签',
  `music_id` bigint(20) DEFAULT NULL COMMENT '音乐ID',
  PRIMARY KEY (`aweme_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频表';

-- ----------------------------
-- Table structure for tb_music
-- ----------------------------
DROP TABLE IF EXISTS `tb_music`;
CREATE TABLE `tb_music` (
  `id` bigint(20) NOT NULL COMMENT '音乐ID',
  `title` varchar(255) DEFAULT NULL COMMENT '标题',
  `author` varchar(255) DEFAULT NULL COMMENT '作者',
  `cover_medium` json DEFAULT NULL COMMENT '中等封面',
  `cover_thumb` json DEFAULT NULL COMMENT '缩略图封面',
  `play_url` json DEFAULT NULL COMMENT '播放地址',
  `duration` int(11) DEFAULT NULL COMMENT '时长',
  `user_count` int(11) DEFAULT NULL COMMENT '使用人数',
  `owner_nickname` varchar(100) DEFAULT NULL COMMENT '拥有者昵称',
  `is_original` tinyint(1) DEFAULT NULL COMMENT '是否原创',
  `owner_id` varchar(100) DEFAULT NULL COMMENT '拥有者ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='音乐表';

-- ----------------------------
-- Table structure for tb_source
-- ----------------------------
DROP TABLE IF EXISTS `tb_source`;
CREATE TABLE `tb_source` (
  `id` bigint(20) NOT NULL COMMENT '资源ID',
  `play_addr` json DEFAULT NULL COMMENT '播放地址',
  `cover` json DEFAULT NULL COMMENT '封面',
  `poster` json DEFAULT NULL COMMENT '海报',
  `height` int(11) DEFAULT NULL COMMENT '高度',
  `width` int(11) DEFAULT NULL COMMENT '宽度',
  `ratio` varchar(20) DEFAULT NULL COMMENT '比例',
  `use_static_cover` tinyint(1) DEFAULT NULL COMMENT '是否使用静态封面',
  `duration` int(11) DEFAULT NULL COMMENT '时长',
  `horizontal_type` int(11) DEFAULT NULL COMMENT '横屏类型',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资源表';

-- ----------------------------
-- Table structure for tb_statistics
-- ----------------------------
DROP TABLE IF EXISTS `tb_statistics`;
CREATE TABLE `tb_statistics` (
  `id` bigint(20) NOT NULL COMMENT '统计ID',
  `admire_count` int(11) DEFAULT NULL COMMENT '赞赏数',
  `comment_count` int(11) DEFAULT NULL COMMENT '评论数',
  `digg_count` int(11) DEFAULT NULL COMMENT '点赞数',
  `collect_count` int(11) DEFAULT NULL COMMENT '收藏数',
  `play_count` int(11) DEFAULT NULL COMMENT '播放数',
  `share_count` int(11) DEFAULT NULL COMMENT '分享数',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统计表';

-- ----------------------------
-- Table structure for tb_collects
-- ----------------------------
DROP TABLE IF EXISTS `tb_collects`;
CREATE TABLE `tb_collects` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '收藏ID',
  `uid` bigint(20) DEFAULT NULL COMMENT '用户ID',
  `aweme_id` bigint(20) DEFAULT NULL COMMENT '视频ID',
  `create_time` int(11) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收藏表';

-- ----------------------------
-- Table structure for tb_comments
-- ----------------------------
DROP TABLE IF EXISTS `tb_comments`;
CREATE TABLE `tb_comments` (
  `comment_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `create_time` int(11) DEFAULT NULL COMMENT '创建时间',
  `ip_location` varchar(100) DEFAULT NULL COMMENT 'IP归属地',
  `aweme_id` bigint(20) DEFAULT NULL COMMENT '视频ID',
  `content` text COMMENT '内容',
  `is_author_digged` tinyint(1) DEFAULT NULL COMMENT '作者是否点赞',
  `is_folded` tinyint(1) DEFAULT NULL COMMENT '是否折叠',
  `is_hot` tinyint(1) DEFAULT NULL COMMENT '是否热评',
  `user_buried` tinyint(1) DEFAULT NULL COMMENT '用户是否踩',
  `user_digged` int(11) DEFAULT NULL COMMENT '用户是否赞',
  `digg_count` bigint(20) DEFAULT NULL COMMENT '点赞数',
  `user_id` bigint(20) DEFAULT NULL COMMENT '用户ID',
  `sec_uid` text COMMENT '加密UID',
  `short_user_id` bigint(20) DEFAULT NULL COMMENT '短用户ID',
  `user_unique_id` varchar(255) DEFAULT NULL COMMENT '用户唯一ID',
  `user_signature` text COMMENT '用户签名',
  `nickname` varchar(100) DEFAULT NULL COMMENT '昵称',
  `avatar` text COMMENT '头像',
  `sub_comment_count` bigint(20) DEFAULT NULL COMMENT '子评论数',
  `last_modify_ts` bigint(20) DEFAULT NULL COMMENT '最后修改时间戳',
  PRIMARY KEY (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评论表';

-- ----------------------------
-- Table structure for tb_comment_diggs
-- ----------------------------
DROP TABLE IF EXISTS `tb_comment_diggs`;
CREATE TABLE `tb_comment_diggs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '璇勮鐐硅禐ID',
  `uid` bigint(20) DEFAULT NULL COMMENT '鐢ㄦ埛ID',
  `comment_id` bigint(20) DEFAULT NULL COMMENT '璇勮ID',
  `create_time` int(11) DEFAULT NULL COMMENT '鍒涘缓鏃堕棿',
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_uid_comment` (`uid`,`comment_id`),
  KEY `idx_comment_uid` (`comment_id`,`uid`),
  KEY `idx_uid_create_time` (`uid`,`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='璇勮鐐硅禐琛?;

-- ----------------------------
-- Table structure for tb_diggs
-- ----------------------------
DROP TABLE IF EXISTS `tb_diggs`;
CREATE TABLE `tb_diggs` (
  `digg_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
  `uid` bigint(20) DEFAULT NULL COMMENT '用户ID',
  `aweme_id` bigint(20) DEFAULT NULL COMMENT '视频ID',
  `create_time` int(11) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`digg_id`),
  UNIQUE KEY `ux_uid_aweme` (`uid`,`aweme_id`),
  KEY `idx_aweme_uid` (`aweme_id`,`uid`),
  KEY `idx_uid_create_time` (`uid`,`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点赞表';

-- ----------------------------
-- Table structure for tb_shares
-- ----------------------------
DROP TABLE IF EXISTS `tb_shares`;
CREATE TABLE `tb_shares` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '分享ID',
  `src_uid` bigint(20) DEFAULT NULL COMMENT '源用户ID',
  `dst_uid` bigint(20) DEFAULT NULL COMMENT '目标用户ID',
  `aweme_id` bigint(20) DEFAULT NULL COMMENT '视频ID',
  `message` text COMMENT '消息',
  `create_time` int(11) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分享表';

-- ----------------------------
-- Table structure for tb_auth_casbin_rule
-- ----------------------------
DROP TABLE IF EXISTS `tb_auth_casbin_rule`;
CREATE TABLE `tb_auth_casbin_rule` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '规则ID',
  `ptype` varchar(100) DEFAULT NULL COMMENT '策略类型',
  `v0` varchar(100) DEFAULT NULL COMMENT '角色/用户',
  `v1` varchar(100) DEFAULT NULL COMMENT '资源',
  `v2` varchar(100) DEFAULT NULL COMMENT '动作',
  `v3` varchar(100) DEFAULT NULL COMMENT '扩展3',
  `v4` varchar(100) DEFAULT NULL COMMENT '扩展4',
  `v5` varchar(100) DEFAULT NULL COMMENT '扩展5',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tb_auth_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Casbin权限规则表';

-- ----------------------------
-- Table structure for tb_auth_access_tokens
-- ----------------------------
SET FOREIGN_KEY_CHECKS = 1;
DROP TABLE IF EXISTS `tb_auth_access_tokens`;
CREATE TABLE `tb_auth_access_tokens` (
                                         `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '令牌ID',
                                         `uid` bigint(20) NOT NULL COMMENT '用户ID',
                                         `action_name` varchar(50) DEFAULT NULL COMMENT '动作名称',
                                         `token` text COMMENT '令牌内容',
                                         `created_at` bigint(20) DEFAULT NULL COMMENT '创建时间',
                                         `expires_at` bigint(20) DEFAULT NULL COMMENT '过期时间',
                                         `client_ip` varchar(50) DEFAULT NULL COMMENT '客户端IP',
                                         `revoked` tinyint(1) DEFAULT 0 COMMENT '是否撤销',
                                         `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                         PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问令牌表';
