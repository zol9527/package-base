package models

import (
	"time"
)

// UserRoleType 用户角色类型
type UserRoleType int

const (
	UserRoleTypeUser  UserRoleType = 0 // 普通用户
	UserRoleTypeAdmin UserRoleType = 1 // 管理员
)

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusInactive  UserStatus = 0 // 未激活
	UserStatusActive    UserStatus = 1 // 激活
	UserStatusSuspended UserStatus = 2 // 暂停
	UserStatusBanned    UserStatus = 3 // 禁用
)

// UserGender 用户性别枚举
type UserGender int

const (
	UserGenderUnknown UserGender = 0 // 未知
	UserGenderMale    UserGender = 1 // 男性
	UserGenderFemale  UserGender = 2 // 女性
	UserGenderOther   UserGender = 3 // 其他
)

/**
 * Function: User
 *
 * Description: 用户模型，用于管理系统中的用户信息，作为主用户表整合所有通用字段
 *
 * @field ID {string} 用户主键ID, uuid
 * @field UUID {string} 用户短唯一ID, 适合对外展示、短链等场景
 * @field Username {string} 用户名
 * @field Email {string} 邮箱地址
 * @field PasswordHash {string} 密码哈希值
 * @field Phone {string} 手机号码
 * @field Status {UserStatus} 用户状态
 * @field Roles {UserRoleType} 用户角色
 * @field Remark {string} 备注信息
 * @field Nickname {string} 昵称
 * @field AvatarURL {string} 头像URL
 * @field Gender {int} 性别
 * @field Country {string} 国家
 * @field Province {string} 省份
 * @field City {string} 城市
 * @field Language {string} 语言
 * @field LoginCount {int} 登录次数
 * @field LastLoginAt {*time.Time} 最后登录时间
 * @field EmailVerifiedAt {*time.Time} 邮箱验证时间
 * @field PhoneVerifiedAt {*time.Time} 手机验证时间
 * @field CreatedAt {time.Time} 创建时间
 * @field UpdatedAt {time.Time} 更新时间
 */
type User struct {
	ID string `gorm:"primaryKey" json:"id"`
	// Uuid 用户短唯一ID，适合对外展示、短链等场景
	Uuid         string       `gorm:"size:22;uniqueIndex;not null;comment:短UUID" json:"uuid"`
	Username     string       `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Email        *string      `gorm:"size:100;uniqueIndex" json:"email"`
	PasswordHash *string      `gorm:"size:255" json:"-"` // 不在JSON中返回密码
	Phone        *string      `gorm:"size:20;index" json:"phone"`
	Status       UserStatus   `gorm:"default:1;index" json:"status"` // 用户状态: 0-未激活, 1-激活, 2-暂停, 3-禁用
	Roles        UserRoleType `gorm:"default:0" json:"roles"`        // 用户角色: 0-普通用户, 1-管理员

	// 基础信息字段
	Remark      *string    `gorm:"type:text;comment:备注信息" json:"remark"`
	FirstName   *string    `gorm:"size:50;comment:名字" json:"first_name"`
	LastName    *string    `gorm:"size:50;comment:姓氏" json:"last_name"`
	Nickname    *string    `gorm:"size:100;comment:昵称" json:"nickname"`
	AvatarURL   *string    `gorm:"size:500;comment:头像URL" json:"avatar_url"`
	Gender      *int       `gorm:"default:0;comment:性别 0:未知 1:男 2:女 3:其他" json:"gender"`
	DateOfBirth *time.Time `gorm:"comment:出生日期" json:"date_of_birth"`
	Bio         *string    `gorm:"type:text;comment:个人简介" json:"bio"`

	// 地址信息字段
	Country    *string `gorm:"size:50;comment:国家" json:"country"`
	Province   *string `gorm:"size:50;comment:省份" json:"province"`
	City       *string `gorm:"size:50;comment:城市" json:"city"`
	Address    *string `gorm:"size:200;comment:详细地址" json:"address"`
	PostalCode *string `gorm:"size:20;comment:邮政编码" json:"postal_code"`

	// 偏好设置字段
	Language *string `gorm:"size:20;comment:语言" json:"language"`
	Timezone *string `gorm:"size:50;comment:时区" json:"timezone"`

	// 社交信息字段
	Website     *string     `gorm:"size:200;comment:个人网站" json:"website"`
	SocialLinks SocialLinks `gorm:"type:json;serializer:json;comment:社交链接JSON" json:"social_links"`

	// 统计字段
	LoginCount          int `gorm:"default:0;comment:登录次数" json:"login_count"`
	ProfileCompleteness int `gorm:"default:0;comment:资料完整度百分比" json:"profile_completeness"`

	// 设置字段
	Theme                *string              `gorm:"size:20;default:'light';comment:主题设置" json:"theme"`
	NotificationSettings NotificationSettings `gorm:"type:json;serializer:json;comment:通知设置JSON" json:"notification_settings"`
	PrivacySettings      PrivacySettings      `gorm:"type:json;serializer:json;comment:隐私设置JSON" json:"privacy_settings"`

	// 时间字段
	LastActivityAt  *time.Time `json:"last_activity_at" gorm:"comment:最后活动时间"`
	LastLoginAt     *time.Time `json:"last_login_at" gorm:"comment:最后登录时间"`
	EmailVerifiedAt *time.Time `json:"email_verified_at" gorm:"comment:邮箱验证时间"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at" gorm:"comment:手机验证时间"`
	CreatedAt       time.Time  `json:"created_at" gorm:"comment:创建时间;autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"comment:更新时间;autoUpdateTime"`

	// 安全字段
	OtpEnabled       bool        `gorm:"default:false;comment:是否启用OTP" json:"otp_enabled"`
	OtpSecret        *string     `gorm:"size:255;comment:OTP密钥" json:"otp_secret"`
	OtpVerifiedAt    *time.Time  `json:"otp_verified_at" gorm:"comment:OTP验证时间"`
	TwoFactorEnabled bool        `gorm:"default:false;comment:是否启用双因素认证" json:"two_factor_enabled"`
	BackupCodes      BackupCodes `gorm:"type:json;serializer:json;comment:备用验证码JSON" json:"backup_codes"`
}

// TableName 获取表名
//
// 描述: 实现GORM的TableName接口，返回自定义表名
//
// 返回:
//
//	string: 表名
func (u *User) TableName() string {
	return "users"
}

// Getter 方法

// GetID 获取用户ID
func (u *User) GetID() string {
	return u.ID
}

// GetUsername 获取用户名
func (u *User) GetUsername() string {
	return u.Username
}

// GetEmail 获取邮箱
func (u *User) GetEmail() string {
	if u.Email == nil {
		return ""
	}
	return *u.Email
}

// GetPhone 获取手机号
func (u *User) GetPhone() string {
	if u.Phone == nil {
		return ""
	}
	return *u.Phone
}

// GetStatus 获取用户状态
func (u *User) GetStatus() UserStatus {
	return u.Status
}

// GetRoles 获取用户角色
func (u *User) GetRoles() UserRoleType {
	return u.Roles
}

// GetFirstName 获取名字
func (u *User) GetFirstName() string {
	if u.FirstName == nil {
		return ""
	}
	return *u.FirstName
}

// GetLastName 获取姓氏
func (u *User) GetLastName() string {
	if u.LastName == nil {
		return ""
	}
	return *u.LastName
}

// GetNickname 获取昵称
func (u *User) GetNickname() string {
	if u.Nickname == nil {
		return ""
	}
	return *u.Nickname
}

// GetAvatarURL 获取头像URL
func (u *User) GetAvatarURL() string {
	if u.AvatarURL == nil {
		return ""
	}
	return *u.AvatarURL
}

// GetGender 获取性别
func (u *User) GetGender() int {
	if u.Gender == nil {
		return 0
	}
	return *u.Gender
}

// GetDateOfBirth 获取出生日期
func (u *User) GetDateOfBirth() *time.Time {
	return u.DateOfBirth
}

// GetBio 获取个人简介
func (u *User) GetBio() string {
	if u.Bio == nil {
		return ""
	}
	return *u.Bio
}

// GetCountry 获取国家
func (u *User) GetCountry() string {
	if u.Country == nil {
		return ""
	}
	return *u.Country
}

// GetProvince 获取省份
func (u *User) GetProvince() string {
	if u.Province == nil {
		return ""
	}
	return *u.Province
}

// GetCity 获取城市
func (u *User) GetCity() string {
	if u.City == nil {
		return ""
	}
	return *u.City
}

// GetAddress 获取详细地址
func (u *User) GetAddress() string {
	if u.Address == nil {
		return ""
	}
	return *u.Address
}

// GetPostalCode 获取邮政编码
func (u *User) GetPostalCode() string {
	if u.PostalCode == nil {
		return ""
	}
	return *u.PostalCode
}

// GetLanguage 获取语言
func (u *User) GetLanguage() string {
	if u.Language == nil {
		return ""
	}
	return *u.Language
}

// GetTimezone 获取时区
func (u *User) GetTimezone() string {
	if u.Timezone == nil {
		return ""
	}
	return *u.Timezone
}

// GetWebsite 获取个人网站
func (u *User) GetWebsite() string {
	if u.Website == nil {
		return ""
	}
	return *u.Website
}

// GetSocialLinks 获取社交链接
func (u *User) GetSocialLinks() SocialLinks {
	return u.SocialLinks
}

// GetNotificationSettings 获取通知设置
func (u *User) GetNotificationSettings() NotificationSettings {
	return u.NotificationSettings
}

// GetPrivacySettings 获取隐私设置
func (u *User) GetPrivacySettings() PrivacySettings {
	return u.PrivacySettings
}

// GetLoginCount 获取登录次数
func (u *User) GetLoginCount() int {
	return u.LoginCount
}

// GetProfileCompleteness 获取资料完整度
func (u *User) GetProfileCompleteness() int {
	return u.ProfileCompleteness
}

// GetTheme 获取主题设置
func (u *User) GetTheme() string {
	if u.Theme == nil {
		return "light"
	}
	return *u.Theme
}

// GetBackupCodes 获取备用验证码
func (u *User) GetBackupCodes() BackupCodes {
	return u.BackupCodes
}

// GetLastActivityAt 获取最后活动时间
func (u *User) GetLastActivityAt() *time.Time {
	return u.LastActivityAt
}

// GetLastLoginAt 获取最后登录时间
func (u *User) GetLastLoginAt() *time.Time {
	return u.LastLoginAt
}

// GetEmailVerifiedAt 获取邮箱验证时间
func (u *User) GetEmailVerifiedAt() *time.Time {
	return u.EmailVerifiedAt
}

// GetPhoneVerifiedAt 获取手机验证时间
func (u *User) GetPhoneVerifiedAt() *time.Time {
	return u.PhoneVerifiedAt
}

// GetCreatedAt 获取创建时间
func (u *User) GetCreatedAt() time.Time {
	return u.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (u *User) GetUpdatedAt() time.Time {
	return u.UpdatedAt
}

// IsOtpEnabled 是否启用OTP
func (u *User) IsOtpEnabled() bool {
	return u.OtpEnabled
}

// GetOtpSecret 获取OTP密钥
func (u *User) GetOtpSecret() string {
	if u.OtpSecret == nil {
		return ""
	}
	return *u.OtpSecret
}

// GetOtpVerifiedAt 获取OTP验证时间
func (u *User) GetOtpVerifiedAt() *time.Time {
	return u.OtpVerifiedAt
}

// IsTwoFactorEnabled 是否启用双因素认证
func (u *User) IsTwoFactorEnabled() bool {
	return u.TwoFactorEnabled
}

// Setter 方法

// SetID 设置用户ID
func (u *User) SetID(id string) {
	u.ID = id
}

// SetUsername 设置用户名
func (u *User) SetUsername(username string) {
	u.Username = username
}

// SetEmail 设置邮箱
func (u *User) SetEmail(email string) {
	u.Email = &email
}

// SetPhone 设置手机号
func (u *User) SetPhone(phone string) {
	u.Phone = &phone
}

// SetStatus 设置用户状态
func (u *User) SetStatus(status UserStatus) {
	u.Status = status
}

// SetRoles 设置用户角色
func (u *User) SetRoles(roles UserRoleType) {
	u.Roles = roles
}

// SetFirstName 设置名字
func (u *User) SetFirstName(firstName string) {
	u.FirstName = &firstName
}

// SetLastName 设置姓氏
func (u *User) SetLastName(lastName string) {
	u.LastName = &lastName
}

// SetNickname 设置昵称
func (u *User) SetNickname(nickname string) {
	u.Nickname = &nickname
}

// SetAvatarURL 设置头像URL
func (u *User) SetAvatarURL(avatarURL string) {
	u.AvatarURL = &avatarURL
}

// SetGender 设置性别
func (u *User) SetGender(gender int) {
	u.Gender = &gender
}

// SetDateOfBirth 设置出生日期
func (u *User) SetDateOfBirth(dateOfBirth *time.Time) {
	u.DateOfBirth = dateOfBirth
}

// SetBio 设置个人简介
func (u *User) SetBio(bio string) {
	u.Bio = &bio
}

// SetCountry 设置国家
func (u *User) SetCountry(country string) {
	u.Country = &country
}

// SetProvince 设置省份
func (u *User) SetProvince(province string) {
	u.Province = &province
}

// SetCity 设置城市
func (u *User) SetCity(city string) {
	u.City = &city
}

// SetAddress 设置详细地址
func (u *User) SetAddress(address string) {
	u.Address = &address
}

// SetPostalCode 设置邮政编码
func (u *User) SetPostalCode(postalCode string) {
	u.PostalCode = &postalCode
}

// SetLanguage 设置语言
func (u *User) SetLanguage(language string) {
	u.Language = &language
}

// SetTimezone 设置时区
func (u *User) SetTimezone(timezone string) {
	u.Timezone = &timezone
}

// SetWebsite 设置个人网站
func (u *User) SetWebsite(website string) {
	u.Website = &website
}

// SetSocialLinks 设置社交链接
func (u *User) SetSocialLinks(socialLinks SocialLinks) {
	u.SocialLinks = socialLinks
}

// SetLoginCount 设置登录次数
func (u *User) SetLoginCount(loginCount int) {
	u.LoginCount = loginCount
}

// SetProfileCompleteness 设置资料完整度
func (u *User) SetProfileCompleteness(completeness int) {
	u.ProfileCompleteness = completeness
}

// SetTheme 设置主题
func (u *User) SetTheme(theme string) {
	u.Theme = &theme
}

// SetNotificationSettings 设置通知设置
func (u *User) SetNotificationSettings(settings NotificationSettings) {
	u.NotificationSettings = settings
}

// SetPrivacySettings 设置隐私设置
func (u *User) SetPrivacySettings(settings PrivacySettings) {
	u.PrivacySettings = settings
}

// SetLastActivityAt 设置最后活动时间
func (u *User) SetLastActivityAt(t *time.Time) {
	u.LastActivityAt = t
}

// SetLastLoginAt 设置最后登录时间
func (u *User) SetLastLoginAt(t *time.Time) {
	u.LastLoginAt = t
}

// SetEmailVerifiedAt 设置邮箱验证时间
func (u *User) SetEmailVerifiedAt(t *time.Time) {
	u.EmailVerifiedAt = t
}

// SetPhoneVerifiedAt 设置手机验证时间
func (u *User) SetPhoneVerifiedAt(t *time.Time) {
	u.PhoneVerifiedAt = t
}

// SetOtpEnabled 设置是否启用OTP
func (u *User) SetOtpEnabled(enabled bool) {
	u.OtpEnabled = enabled
}

// SetOtpSecret 设置OTP密钥
func (u *User) SetOtpSecret(secret string) {
	u.OtpSecret = &secret
}

// SetOtpVerifiedAt 设置OTP验证时间
func (u *User) SetOtpVerifiedAt(t *time.Time) {
	u.OtpVerifiedAt = t
}

// SetTwoFactorEnabled 设置是否启用双因素认证
func (u *User) SetTwoFactorEnabled(enabled bool) {
	u.TwoFactorEnabled = enabled
}

// SetBackupCodes 设置备用验证码
func (u *User) SetBackupCodes(codes BackupCodes) {
	u.BackupCodes = codes
}

// SetPasswordHash 设置密码哈希
func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = &hash
}

// GetPasswordHash 获取密码哈希
func (u *User) GetPasswordHash() string {
	if u.PasswordHash == nil {
		return ""
	}
	return *u.PasswordHash
}

// SetRemark 设置备注
func (u *User) SetRemark(remark string) {
	u.Remark = &remark
}

// GetRemark 获取备注
func (u *User) GetRemark() string {
	if u.Remark == nil {
		return ""
	}
	return *u.Remark
}

// ========================================
// 🔧 业务方法
// ========================================

// GetFullName 获取完整姓名
//
// 函数名: GetFullName
//
// 描述: 获取用户的完整姓名（姓氏 + 名字）
//
// 返回:
//
//	string: 完整姓名
func (u *User) GetFullName() string {
	firstName := u.GetFirstName()
	lastName := u.GetLastName()

	if firstName == "" && lastName == "" {
		return ""
	}
	if firstName == "" {
		return lastName
	}
	if lastName == "" {
		return firstName
	}
	return lastName + firstName
}

// GetDisplayName 获取显示名称
//
// 函数名: GetDisplayName
//
// 描述: 获取用户的显示名称（优先级：昵称 > 完整姓名 > 用户名）
//
// 返回:
//
//	string: 显示名称
func (u *User) GetDisplayName() string {
	nickname := u.GetNickname()
	if nickname != "" {
		return nickname
	}

	fullName := u.GetFullName()
	if fullName != "" {
		return fullName
	}

	return u.GetUsername()
}

// IsEmailVerified 检查邮箱是否已验证
//
// 函数名: IsEmailVerified
//
// 描述: 检查用户邮箱是否已验证
//
// 返回:
//
//	bool: 是否已验证
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// IsPhoneVerified 检查手机号是否已验证
//
// 函数名: IsPhoneVerified
//
// 描述: 检查用户手机号是否已验证
//
// 返回:
//
//	bool: 是否已验证
func (u *User) IsPhoneVerified() bool {
	return u.PhoneVerifiedAt != nil
}

// GetAge 根据出生日期计算年龄
//
// 函数名: GetAge
//
// 描述: 根据出生日期计算用户年龄
//
// 返回:
//
//	int: 年龄，如果未设置出生日期返回-1
func (u *User) GetAge() int {
	if u.DateOfBirth == nil {
		return -1
	}

	now := time.Now()
	age := now.Year() - u.DateOfBirth.Year()

	// 如果今年的生日还没到，年龄减1
	if now.YearDay() < u.DateOfBirth.YearDay() {
		age--
	}

	return age
}

// IsAdult 判断是否成年
//
// 函数名: IsAdult
//
// 描述: 判断用户是否成年（18岁及以上）
//
// 返回:
//
//	bool: 是否成年，如果未设置出生日期返回false
func (u *User) IsAdult() bool {
	age := u.GetAge()
	return age >= 18
}

// GetLocalTime 获取用户时区的本地时间
//
// 函数名: GetLocalTime
//
// 描述: 根据用户设置的时区获取本地时间
//
// 返回:
//
//	time.Time: 本地时间
//	error: 错误信息
func (u *User) GetLocalTime() (time.Time, error) {
	timezone := u.GetTimezone()
	if timezone == "" {
		// 默认使用UTC时间
		return time.Now().UTC(), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().In(loc), nil
}

// UpdateProfileCompleteness 更新资料完整度
//
// 函数名: UpdateProfileCompleteness
//
// 描述: 计算并更新用户资料的完整度百分比
func (u *User) UpdateProfileCompleteness() {
	totalFields := 20 // 总字段数
	completedFields := 0

	// 必填字段
	if u.Username != "" {
		completedFields++
	}

	// 可选字段
	if u.Email != nil && *u.Email != "" {
		completedFields++
	}
	if u.Phone != nil && *u.Phone != "" {
		completedFields++
	}
	if u.FirstName != nil && *u.FirstName != "" {
		completedFields++
	}
	if u.LastName != nil && *u.LastName != "" {
		completedFields++
	}
	if u.Nickname != nil && *u.Nickname != "" {
		completedFields++
	}
	if u.AvatarURL != nil && *u.AvatarURL != "" {
		completedFields++
	}
	if u.Gender != nil {
		completedFields++
	}
	if u.DateOfBirth != nil {
		completedFields++
	}
	if u.Bio != nil && *u.Bio != "" {
		completedFields++
	}
	if u.Country != nil && *u.Country != "" {
		completedFields++
	}
	if u.Province != nil && *u.Province != "" {
		completedFields++
	}
	if u.City != nil && *u.City != "" {
		completedFields++
	}
	if u.Address != nil && *u.Address != "" {
		completedFields++
	}
	if u.PostalCode != nil && *u.PostalCode != "" {
		completedFields++
	}
	if u.Language != nil && *u.Language != "" {
		completedFields++
	}
	if u.Timezone != nil && *u.Timezone != "" {
		completedFields++
	}
	if u.Website != nil && *u.Website != "" {
		completedFields++
	}
	// struct 判空：只要有任一字段非空即视为已填写
	if u.SocialLinks.Github != "" || u.SocialLinks.Twitter != "" || u.SocialLinks.Wechat != "" || u.SocialLinks.Website != "" {
		completedFields++
	}
	if u.Theme != nil && *u.Theme != "" {
		completedFields++
	}

	u.ProfileCompleteness = (completedFields * 100) / totalFields
}

// UpdateLastActivity 更新最后活动时间
//
// 函数名: UpdateLastActivity
//
// 描述: 更新用户的最后活动时间为当前时间
func (u *User) UpdateLastActivity() {
	now := time.Now()
	u.LastActivityAt = &now
}

// GetDisplayRole 获取角色显示名称
//
// 函数名: GetDisplayRole
//
// 描述: 获取用户角色的中文显示名称
//
// 返回:
//
//	string: 角色显示名称
func (u *User) GetDisplayRoleInZh() string {
	switch u.Roles {
	case UserRoleTypeUser:
		return "普通用户"
	case UserRoleTypeAdmin:
		return "管理员"
	default:
		return "未知角色"
	}
}

// GetDisplayRoleInEn 获取角色字符串
//
// 函数名: GetDisplayRoleInEn
//
// 描述: 获取用户角色的英文显示名称
//
// 返回:
//
//	string: 角色字符串
func (u *User) GetDisplayRoleInEn() string {
	switch u.Roles {
	case UserRoleTypeUser:
		return "user"
	case UserRoleTypeAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// GetDisplayStatus 获取状态显示名称
//
// 函数名: GetDisplayStatus
//
// 描述: 获取用户状态的中文显示名称
//
// 返回:
//
//	string: 状态显示名称
func (u *User) GetDisplayStatusInZh() string {
	switch u.Status {
	case UserStatusInactive:
		return "未激活"
	case UserStatusActive:
		return "正常"
	case UserStatusSuspended:
		return "暂停"
	case UserStatusBanned:
		return "禁用"
	default:
		return "未知状态"
	}
}

// GetDisplayGender 获取性别显示名称
//
// 函数名: GetDisplayGender
//
// 描述: 获取用户性别的中文显示名称
//
// 返回:
//
//	string: 性别显示名称
func (u *User) GetDisplayGenderInZh() string {
	if u.Gender == nil {
		return "未知"
	}

	switch *u.Gender {
	case 0:
		return "未知"
	case 1:
		return "男"
	case 2:
		return "女"
	case 3:
		return "其他"
	default:
		return "未知"
	}
}

// IsActive 检查用户是否激活
//
// 函数名: IsActive
//
// 描述: 检查用户是否处于激活状态
//
// 返回:
//
//	bool: 是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsAdmin 检查是否为管理员
//
// 函数名: IsAdmin
//
// 描述: 检查用户是否为管理员
//
// 返回:
//
//	bool: 是否为管理员
func (u *User) IsAdmin() bool {
	return u.Roles == UserRoleTypeAdmin
}

// IsNewUser 检查是否为新用户
//
// 函数名: IsNewUser
//
// 描述: 检查用户是否为新用户（登录次数为0）
//
// 返回:
//
//	bool: 是否为新用户
func (u *User) IsNewUser() bool {
	return u.LoginCount == 0
}

// SetGenderToNull 设置性别为NULL
//
// 描述: 将性别字段设置为NULL
func (u *User) SetGenderToNull() {
	u.Gender = nil
}

// SetEmailToNull 设置邮箱为NULL
//
// 描述: 将邮箱字段设置为NULL
func (u *User) SetEmailToNull() {
	u.Email = nil
}

// SetPasswordHashToNull 设置密码哈希为NULL
//
// 描述: 将密码哈希字段设置为NULL
func (u *User) SetPasswordHashToNull() {
	u.PasswordHash = nil
}

// SetPhoneToNull 设置电话为NULL
//
// 描述: 将电话字段设置为NULL
func (u *User) SetPhoneToNull() {
	u.Phone = nil
}

// SetRemarkToNull 设置备注为NULL
//
// 描述: 将备注字段设置为NULL
func (u *User) SetRemarkToNull() {
	u.Remark = nil
}

// SetNicknameToNull 设置昵称为NULL
//
// 描述: 将昵称字段设置为NULL
func (u *User) SetNicknameToNull() {
	u.Nickname = nil
}

// SetAvatarURLToNull 设置头像URL为NULL
//
// 描述: 将头像URL字段设置为NULL
func (u *User) SetAvatarURLToNull() {
	u.AvatarURL = nil
}

// SetCountryToNull 设置国家为NULL
//
// 描述: 将国家字段设置为NULL
func (u *User) SetCountryToNull() {
	u.Country = nil
}

// SetProvinceToNull 设置省份为NULL
//
// 描述: 将省份字段设置为NULL
func (u *User) SetProvinceToNull() {
	u.Province = nil
}

// SetCityToNull 设置城市为NULL
//
// 描述: 将城市字段设置为NULL
func (u *User) SetCityToNull() {
	u.City = nil
}

// SetLanguageToNull 设置语言为NULL
//
// 描述: 将语言字段设置为NULL
func (u *User) SetLanguageToNull() {
	u.Language = nil
}

// HasEmail 检查是否有邮箱
//
// 描述: 检查用户是否设置了邮箱
//
// 返回:
//
//	bool: 是否有邮箱
func (u *User) HasEmail() bool {
	return u.Email != nil && *u.Email != ""
}

// HasPhone 检查是否有电话
//
// 描述: 检查用户是否设置了电话号码
//
// 返回:
//
//	bool: 是否有电话
func (u *User) HasPhone() bool {
	return u.Phone != nil && *u.Phone != ""
}

// HasNickname 检查是否有昵称
//
// 描述: 检查用户是否设置了昵称
//
// 返回:
//
//	bool: 是否有昵称
func (u *User) HasNickname() bool {
	return u.Nickname != nil && *u.Nickname != ""
}

// HasAvatarURL 检查是否有头像
//
// 描述: 检查用户是否设置了头像URL
//
// 返回:
//
//	bool: 是否有头像
func (u *User) HasAvatarURL() bool {
	return u.AvatarURL != nil && *u.AvatarURL != ""
}

// HasRemark 检查是否有备注
//
// 描述: 检查用户是否设置了备注
//
// 返回:
//
//	bool: 是否有备注
func (u *User) HasRemark() bool {
	return u.Remark != nil && *u.Remark != ""
}

// SocialLinks 社交链接结构体
//
// Description: 用户社交账号信息，序列化为 JSON 存储
type SocialLinks struct {
	Github  string `json:"github,omitempty"`
	Twitter string `json:"twitter,omitempty"`
	Wechat  string `json:"wechat,omitempty"`
	Website string `json:"website,omitempty"`
}

// NotificationSettings 通知设置结构体
//
// Description: 用户通知偏好设置，序列化为 JSON 存储
type NotificationSettings struct {
	Email bool `json:"email"`
	SMS   bool `json:"sms"`
	Push  bool `json:"push"`
}

// PrivacySettings 隐私设置结构体
//
// Description: 用户隐私设置，序列化为 JSON 存储
type PrivacySettings struct {
	ProfileVisible bool `json:"profile_visible"`
	Searchable     bool `json:"searchable"`
}

// BackupCodes 备用验证码结构体
//
// Description: 用户二步验证备用码，序列化为 JSON 存储
type BackupCodes struct {
	Codes []string `json:"codes"`
}
