/*
 Navicat Premium Dump SQL

 Source Server         : vm
 Source Server Type    : MySQL
 Source Server Version : 80042 (8.0.42)
 Source Host           : 192.168.101.80:3306
 Source Schema         : game_db

 Target Server Type    : MySQL
 Target Server Version : 80042 (8.0.42)
 File Encoding         : 65001

 Date: 10/05/2025 17:04:42
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for chat_messages
-- ----------------------------
DROP TABLE IF EXISTS `chat_messages`;
CREATE TABLE `chat_messages`  (
  `id` bigint NOT NULL,
  `channel` int NOT NULL,
  `sender_id` bigint NOT NULL,
  `sender_name` varchar(255) NOT NULL,
  `receiver_id` bigint NOT NULL,
  `content` varchar(255) NOT NULL,
  `send_time` int NOT NULL,
  `extra_data` varchar(255) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_sender_id` (`sender_id`),
  INDEX `idx_receiver_id` (`receiver_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for friend_requests
-- ----------------------------
DROP TABLE IF EXISTS `friend_requests`;
CREATE TABLE `friend_requests`  (
  `id` bigint NOT NULL,
  `from_user_id` bigint NOT NULL,
  `to_user_id` bigint NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_from_user_id` (`from_user_id`),
  INDEX `idx_to_user_id` (`to_user_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for friends
-- ----------------------------
DROP TABLE IF EXISTS `friends`;
CREATE TABLE `friends`  (
  `id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `friend_id` bigint NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_friend_id` (`friend_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for guild_applications
-- ----------------------------
DROP TABLE IF EXISTS `guild_applications`;
CREATE TABLE `guild_applications`  (
  `id` bigint NOT NULL,
  `guild_id` bigint NULL DEFAULT NULL,
  `user_id` bigint NULL DEFAULT NULL,
  `apply_time` datetime NULL DEFAULT NULL,
  `status` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_guild_id` (`guild_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for guild_invitations
-- ----------------------------
DROP TABLE IF EXISTS `guild_invitations`;
CREATE TABLE `guild_invitations`  (
  `id` bigint NOT NULL,
  `guild_id` bigint NOT NULL,
  `inviter_id` bigint NOT NULL,
  `invitee_id` bigint NOT NULL,
  `status` int NOT NULL,
  `created_at` datetime NOT NULL,
  `expire_time` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_guild_id` (`guild_id`),
  INDEX `idx_inviter_id` (`inviter_id`),
  INDEX `idx_invitee_id` (`invitee_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for guild_members
-- ----------------------------
DROP TABLE IF EXISTS `guild_members`;
CREATE TABLE `guild_members`  (
  `id` bigint NOT NULL,
  `guild_id` bigint NULL DEFAULT NULL,
  `user_id` bigint NULL DEFAULT NULL,
  `role` int NULL DEFAULT NULL,
  `join_time` datetime NULL DEFAULT NULL,
  `last_login` datetime NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_guild_id` (`guild_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for guilds
-- ----------------------------
DROP TABLE IF EXISTS `guilds`;
CREATE TABLE `guilds`  (
  `id` bigint NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `announcement` varchar(255) NOT NULL,
  `master_id` bigint NOT NULL,
  `created_at` datetime NOT NULL,
  `max_members` int NOT NULL,
  `version` int NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` bigint NOT NULL,
  `level` int NOT NULL DEFAULT 1,
  `exp` int NOT NULL DEFAULT 0,
  `username` varchar(50) NOT NULL,
  `email` varchar(30) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `salt` varchar(255) NOT NULL,
  `role` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_username` (`username`),
  INDEX `idx_email` (`email`),
  INDEX `idx_role` (`role`),
  INDEX `idx_created_at` (`created_at`),
  INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for inventory_items
-- ----------------------------
DROP TABLE IF EXISTS `inventory_items`;
CREATE TABLE IF NOT EXISTS inventory_items (
    id BIGINT NOT NULL,          -- 物品实例ID（雪花算法生成）
    user_id BIGINT NOT NULL,                  -- 用户ID
    template_id BIGINT NOT NULL,              -- 策划配置表模板ID
    count INT NOT NULL DEFAULT 1,             -- 数量
    equipped BOOLEAN NOT NULL DEFAULT FALSE,  -- 是否已装备
    created_at BIGINT NOT NULL,               -- 创建时间
    updated_at BIGINT NOT NULL,               -- 更新时间
    PRIMARY KEY (`id`) USING BTREE,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_template_id` (`template_id`),
    INDEX `idx_user_template` (`user_id`, `template_id`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for equipments
-- ----------------------------
DROP TABLE IF EXISTS `equipments`;
CREATE TABLE IF NOT EXISTS equipments (
    id BIGINT NOT NULL,                       -- 装备实例ID（雪花算法生成）
    user_id BIGINT NOT NULL,                  -- 用户ID
    template_id BIGINT NOT NULL,              -- 策划配置表模板ID
    slot INT NOT NULL,                        -- 装备槽位
    created_at BIGINT NOT NULL,               -- 创建时间
    updated_at BIGINT NOT NULL,               -- 更新时间
    PRIMARY KEY (`id`) USING BTREE,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_user_slot` (`user_id`, `slot`),
    INDEX `idx_template_id` (`template_id`),
    UNIQUE KEY `uk_user_slot` (`user_id`, `slot`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user_cards
-- ----------------------------
DROP TABLE IF EXISTS `user_cards`;
CREATE TABLE IF NOT EXISTS user_cards (
    id BIGINT NOT NULL,                       -- 卡牌实例ID（雪花算法生成）
    user_id BIGINT NOT NULL,                  -- 用户ID
    template_id BIGINT NOT NULL,              -- 策划配置表模板ID
    level INT NOT NULL DEFAULT 1,             -- 卡牌等级
    star INT NOT NULL DEFAULT 0,              -- 卡牌星级
    created_at BIGINT NOT NULL,               -- 创建时间
    updated_at BIGINT NOT NULL,               -- 更新时间
    PRIMARY KEY (`id`) USING BTREE,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_template_id` (`template_id`),
    INDEX `idx_user_template` (`user_id`, `template_id`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for pets
-- ----------------------------
DROP TABLE IF EXISTS `pets`;
CREATE TABLE IF NOT EXISTS pets (
    id BIGINT NOT NULL,                       -- 宠物实例ID（雪花算法生成）
    user_id BIGINT NOT NULL,                  -- 用户ID
    template_id BIGINT NOT NULL,              -- 策划配置表模板ID
    name VARCHAR(255) NOT NULL,               -- 宠物名称
    level INT NOT NULL DEFAULT 1,             -- 宠物等级
    exp INT NOT NULL DEFAULT 0,               -- 当前经验
    is_battle BOOLEAN NOT NULL DEFAULT FALSE, -- 是否出战
    created_at DATETIME NOT NULL,             -- 创建时间
    updated_at DATETIME NOT NULL,             -- 更新时间
    PRIMARY KEY (`id`) USING BTREE,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_template_id` (`template_id`),
    INDEX `idx_user_template` (`user_id`, `template_id`),
    INDEX `idx_is_battle` (`is_battle`),
    INDEX `idx_user_battle` (`user_id`, `is_battle`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for monthly_signs
-- ----------------------------
DROP TABLE IF EXISTS `monthly_signs`;
CREATE TABLE IF NOT EXISTS monthly_signs (
    user_id BIGINT NOT NULL,                  -- 用户ID（主键）
    year INT NOT NULL,                         -- 年份
    month INT NOT NULL,                        -- 月份
    sign_days INT NOT NULL DEFAULT 0,          -- 已签到的日期位图（bitmap）
    last_sign_time DATETIME NOT NULL,          -- 最后签到时间
    created_at DATETIME NOT NULL,              -- 创建时间
    updated_at DATETIME NOT NULL,              -- 更新时间
    PRIMARY KEY (`user_id`) USING BTREE,
    INDEX `idx_year_month` (`year`, `month`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for monthly_sign_rewards
-- ----------------------------
DROP TABLE IF EXISTS `monthly_sign_rewards`;
CREATE TABLE IF NOT EXISTS monthly_sign_rewards (
    user_id BIGINT NOT NULL,                  -- 用户ID（主键）
    year INT NOT NULL,                         -- 年份
    month INT NOT NULL,                        -- 月份
    reward_days INT NOT NULL DEFAULT 0,        -- 已领取奖励的累计天数位图（bitmap）
    created_at DATETIME NOT NULL,              -- 创建时间
    updated_at DATETIME NOT NULL,              -- 更新时间
    PRIMARY KEY (`user_id`) USING BTREE,
    INDEX `idx_year_month` (`year`, `month`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
