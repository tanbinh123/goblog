package provider

import (
	"errors"
	"fmt"
	"github.com/medivhzhan/weapp/v3"
	"gorm.io/gorm"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/request"
	"regexp"
	"strings"
	"time"
)

func (w *Website) GetUserList(ops func(tx *gorm.DB) *gorm.DB, page, pageSize int) ([]*model.User, int64) {
	var users []*model.User
	var total int64
	offset := (page - 1) * pageSize
	tx := w.DB.Model(&model.User{})
	if ops != nil {
		tx = ops(tx)
	} else {
		tx = tx.Order("id desc")
	}
	tx.Count(&total).Limit(pageSize).Offset(offset).Find(&users)
	if len(users) > 0 {
		groups := w.GetUserGroups()
		for i := range users {
			users[i].GetThumb(w.PluginStorage.StorageUrl)
			for g := range groups {
				if users[i].GroupId == groups[g].Id {
					users[i].Group = groups[g]
				}
			}
		}
	}

	return users, total
}

func (w *Website) GetUserInfoById(userId uint) (*model.User, error) {
	var user model.User
	err := w.DB.Where("`id` = ?", userId).Take(&user).Error

	if err != nil {
		return nil, err
	}
	user.GetThumb(w.PluginStorage.StorageUrl)
	return &user, nil
}

func (w *Website) GetUserInfoByUserName(userName string) (*model.User, error) {
	var user model.User
	err := w.DB.Where("`user_name` = ?", userName).Take(&user).Error

	if err != nil {
		return nil, err
	}
	user.GetThumb(w.PluginStorage.StorageUrl)
	return &user, nil
}

func (w *Website) GetUserInfoByEmail(email string) (*model.User, error) {
	var user model.User
	err := w.DB.Where("`email` = ?", email).Take(&user).Error

	if err != nil {
		return nil, err
	}
	user.GetThumb(w.PluginStorage.StorageUrl)
	return &user, nil
}

func (w *Website) GetUserInfoByPhone(phone string) (*model.User, error) {
	var user model.User
	err := w.DB.Where("`phone` = ?", phone).Take(&user).Error

	if err != nil {
		return nil, err
	}
	user.GetThumb(w.PluginStorage.StorageUrl)
	return &user, nil
}

func (w *Website) CheckUserInviteCode(inviteCode string) (*model.User, error) {
	var user model.User
	err := w.DB.Where("`invite_code` = ?", inviteCode).Take(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (w *Website) GetUsersInfoByIds(userIds []uint) []*model.User {
	var users []*model.User
	if len(userIds) == 0 {
		return users
	}
	w.DB.Where("`id` IN(?)", userIds).Find(&users)

	return users
}

func (w *Website) SaveUserInfo(req *request.UserRequest) error {
	var user = model.User{
		UserName:   req.UserName,
		RealName:   req.RealName,
		AvatarURL:  req.AvatarURL,
		Phone:      req.Phone,
		Email:      req.Email,
		IsRetailer: req.IsRetailer,
		ParentId:   req.ParentId,
		InviteCode: req.InviteCode,
		GroupId:    req.GroupId,
		ExpireTime: req.ExpireTime,
		Status:     req.Status,
	}
	req.Password = strings.TrimSpace(req.Password)
	if req.Password != "" {
		user.EncryptPassword(req.Password)
	}
	if user.GroupId == 0 {
		user.GroupId = w.PluginUser.DefaultGroupId
	}
	if req.Id > 0 {
		_, err := w.GetUserInfoById(req.Id)
		if err != nil {
			// 用户不存在
			return err
		}
		user.Id = req.Id
	}
	err := w.DB.Save(&user).Error

	return err
}

func (w *Website) DeleteUserInfo(userId uint) error {
	var user model.User
	err := w.DB.Where("`id` = ?", userId).Take(&user).Error

	if err != nil {
		return err
	}

	err = w.DB.Delete(&user).Error

	return err
}

func (w *Website) GetUserGroups() []*model.UserGroup {
	var groups []*model.UserGroup

	w.DB.Order("level asc,id asc").Find(&groups)

	return groups
}

func (w *Website) GetUserGroupInfo(groupId uint) (*model.UserGroup, error) {
	var group model.UserGroup

	err := w.DB.Where("`id` = ?", groupId).Take(&group).Error

	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (w *Website) SaveUserGroupInfo(req *request.UserGroupRequest) error {
	var group = model.UserGroup{
		Title:       req.Title,
		Description: req.Description,
		Level:       req.Level,
		Price:       req.Price,
		Status:      1,
		Setting:     req.Setting,
	}
	if req.Id > 0 {
		_, err := w.GetUserGroupInfo(req.Id)
		if err != nil {
			// 不存在
			return err
		}
		group.Id = req.Id
	}
	err := w.DB.Save(&group).Error

	return err
}

func (w *Website) GetUserWechatByOpenid(openid string) (*model.UserWechat, error) {
	var userWechat model.UserWechat
	if err := w.DB.Where("`openid` = ?", openid).First(&userWechat).Error; err != nil {
		return nil, err
	}

	return &userWechat, nil
}

func (w *Website) GetUserByUnionId(unionId string) (*model.User, error) {
	var userWechat model.UserWechat
	if err := w.DB.Where("`union_id` = ? AND user_id > 0", unionId).First(&userWechat).Error; err != nil {
		return nil, err
	}

	user, err := w.GetUserInfoById(userWechat.UserId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (w *Website) GetUserWechatByUserId(userId uint) (*model.UserWechat, error) {
	var userWechat model.UserWechat
	if err := w.DB.Where("`user_id` = ?", userId).First(&userWechat).Error; err != nil {
		return nil, err
	}

	return &userWechat, nil
}

func (w *Website) DeleteUserGroup(groupId uint) error {
	var group model.UserGroup
	err := w.DB.Where("`id` = ?", groupId).Take(&group).Error

	if err != nil {
		return err
	}

	err = w.DB.Delete(&group).Error

	return err
}

func (w *Website) RegisterUser(req *request.ApiRegisterRequest) (*model.User, error) {
	if req.UserName == "" || req.Password == "" {
		return nil, errors.New(w.Lang("请正确填写用户名和密码"))
	}
	if len(req.Password) < 6 {
		return nil, errors.New(w.Lang("请输入6位以上的密码"))
	}
	_, err := w.GetUserInfoByUserName(req.UserName)
	if err == nil {
		return nil, errors.New(w.Lang("该用户名已被注册"))
	}
	if req.Phone != "" {
		if !w.VerifyCellphoneFormat(req.Phone) {
			return nil, errors.New(w.Lang("手机号不正确"))
		}
		_, err := w.GetUserInfoByPhone(req.Phone)
		if err == nil {
			return nil, errors.New(w.Lang("该手机号已被注册"))
		}
	}
	if req.Email != "" {
		if !w.VerifyEmailFormat(req.Email) {
			return nil, errors.New(w.Lang("邮箱不正确"))
		}
		_, err := w.GetUserInfoByEmail(req.Email)
		if err == nil {
			return nil, errors.New(w.Lang("该邮箱已被注册"))
		}
	}
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New(w.Lang("邮箱和手机号至少填写一个"))
	}

	user := model.User{
		UserName:  req.UserName,
		RealName:  req.RealName,
		AvatarURL: req.AvatarURL,
		ParentId:  req.InviteId,
		Phone:     req.Phone,
		Email:     req.Email,
		GroupId:   w.PluginUser.DefaultGroupId,
		Status:    1,
	}
	user.EncryptPassword(req.Password)
	w.DB.Save(&user)

	_ = user.EncodeToken(w.DB)

	return &user, nil
}

func (w *Website) LoginViaWeapp(req *request.ApiLoginRequest) (*model.User, error) {
	loginRs, err := w.GetWeappClient(false).Login(req.Code)
	if err != nil {
		return nil, err
	}
	//log.Printf("%#v", loginRs)
	if loginRs.OpenID == "" {
		//openid 不在？
		return nil, errors.New("无法获取openid")
	}

	var wecahtUserInfo *weapp.UserInfo
	wecahtUserInfo, err = w.GetWeappClient(false).DecryptUserInfo(loginRs.SessionKey, req.RawData, req.EncryptedData, req.Signature, req.Iv)
	if err != nil {
		wecahtUserInfo = &weapp.UserInfo{
			Avatar:   req.Avatar,
			Gender:   int(req.Gender),
			Country:  req.County,
			City:     req.City,
			Language: "",
			Nickname: req.NickName,
			Province: req.Province,
		}
	}
	// 拿到openid
	userWechat, userErr := w.GetUserWechatByOpenid(loginRs.OpenID)
	var user *model.User
	if userErr != nil {
		//系统没记录，则插入一条记录
		user = &model.User{
			UserName:  wecahtUserInfo.Nickname,
			AvatarURL: wecahtUserInfo.Avatar,
			ParentId:  req.InviteId,
			GroupId:   w.PluginUser.DefaultGroupId,
			Status:    1,
		}

		err = w.DB.Save(user).Error
		if err != nil {
			return nil, err
		}

		userWechat = &model.UserWechat{
			UserId:    user.Id,
			Nickname:  wecahtUserInfo.Nickname,
			AvatarURL: wecahtUserInfo.Avatar,
			Gender:    wecahtUserInfo.Gender,
			Openid:    loginRs.OpenID,
			UnionId:   loginRs.UnionID,
			Platform:  config.PlatformWeapp,
			Status:    1,
		}

		err = w.DB.Save(userWechat).Error
		if err != nil {
			//删掉
			w.DB.Delete(user)
			return nil, err
		}

		go w.DownloadAvatar(userWechat.AvatarURL, user)
	} else {
		user, err = w.GetUserInfoById(userWechat.UserId)
		if err != nil {
			return nil, err
		}
		//更新信息
		if wecahtUserInfo.Nickname != "" && (userWechat.Nickname != wecahtUserInfo.Nickname || userWechat.AvatarURL != wecahtUserInfo.Avatar) {
			user.UserName = wecahtUserInfo.Nickname
			user.AvatarURL = wecahtUserInfo.Avatar
			err = w.DB.Save(user).Error
			if err != nil {
				return nil, err
			}

			userWechat.Nickname = wecahtUserInfo.Nickname
			userWechat.AvatarURL = wecahtUserInfo.Avatar
			err = w.DB.Save(userWechat).Error
			if err != nil {
				return nil, err
			}
		}
	}

	_ = user.EncodeToken(w.DB)

	return user, nil
}

func (w *Website) LoginViaWechat(req *request.ApiLoginRequest) (*model.User, error) {
	openid := library.CodeCache.GetByCode(req.Code, false)
	if openid == "" {
		return nil, errors.New(w.Lang("验证码不正确"))
	}
	// auto register
	userWechat, err := w.GetUserWechatByOpenid(openid)
	if err != nil {
		return nil, errors.New(w.Lang("用户信息不完整"))
	}
	var user *model.User
	if userWechat.UserId == 0 {
		user = &model.User{
			UserName:  userWechat.Nickname,
			AvatarURL: userWechat.AvatarURL,
			GroupId:   w.PluginUser.DefaultGroupId,
			Password:  "",
			Status:    1,
		}
		w.DB.Save(user)
		userWechat.UserId = user.Id
		w.DB.Save(userWechat)
	} else {
		user, err = w.GetUserInfoById(userWechat.UserId)
		if err != nil {
			return nil, errors.New(w.Lang("用户信息不完整"))
		}
	}
	if req.InviteId > 0 && user.ParentId == 0 {
		user.ParentId = req.InviteId
		w.DB.Save(user)
	}

	_ = user.EncodeToken(w.DB)

	return user, nil
}

func (w *Website) LoginViaPassword(req *request.ApiLoginRequest) (*model.User, error) {
	var user model.User
	if w.VerifyEmailFormat(req.UserName) {
		//邮箱登录
		err := w.DB.Where("email = ?", req.UserName).First(&user).Error
		if err != nil {
			return nil, err
		}
	} else if w.VerifyCellphoneFormat(req.UserName) {
		//手机号登录
		err := w.DB.Where("phone = ?", req.UserName).First(&user).Error
		if err != nil {
			return nil, err
		}
	} else {
		//用户名登录
		err := w.DB.Where("user_name = ?", req.UserName).First(&user).Error
		if err != nil {
			return nil, err
		}
	}
	//验证密码
	ok := user.CheckPassword(req.Password)
	if !ok {
		return nil, errors.New(w.Lang("密码错误"))
	}

	_ = user.EncodeToken(w.DB)

	return &user, nil
}

func (w *Website) VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func (w *Website) VerifyCellphoneFormat(cellphone string) bool {
	pattern := `1[3-9][0-9]{9}` //宽匹配手机号
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(cellphone)
}

func (w *Website) DownloadAvatar(avatarUrl string, userInfo *model.User) {
	if avatarUrl == "" || !strings.HasPrefix(avatarUrl, "http") {
		return
	}

	//生成用户文件
	tmpName := fmt.Sprintf("%010d.jpg", userInfo.Id)
	filePath := fmt.Sprintf("/uploads/avatar/%s/%s/%s", tmpName[:3], tmpName[3:6], tmpName[6:])
	attach, err := w.DownloadRemoteImage(avatarUrl, filePath)
	if err != nil {
		return
	}
	//写入完成，更新数据库
	userInfo.AvatarURL = attach.FileLocation
	w.DB.Model(userInfo).UpdateColumn("avatar_url", userInfo.AvatarURL)
}

func (w *Website) GetRetailerMembers(retailerId uint, page, pageSize int) ([]*model.User, int64) {
	var users []*model.User
	var total int64
	offset := (page - 1) * pageSize
	tx := w.DB.Model(&model.User{}).Where("`parent_id` = ?", retailerId)
	tx.Count(&total).Order("id desc").Limit(pageSize).Offset(offset).Find(&users)

	return users, total
}

func (w *Website) UpdateUserRealName(userId uint, realName string) error {
	err := w.DB.Model(&model.User{}).Where("`id` = ?", userId).UpdateColumn("real_name", realName).Error

	return err
}

func (w *Website) SetRetailerInfo(userId uint, isRetailer int) error {
	err := w.DB.Model(&model.User{}).Where("`id` = ?", userId).UpdateColumn("is_retailer", isRetailer).Error

	return err
}

func (w *Website) UpdateUserInfo(userId uint, req *request.UserRequest) error {
	user, err := w.GetUserInfoById(userId)
	if err != nil {
		return err
	}

	exist, err := w.GetUserInfoByUserName(req.UserName)
	if err == nil && exist.Id != user.Id {
		return errors.New(w.Lang("该用户名已被注册"))
	}

	if user.Phone != "" {
		req.Phone = ""
	}
	if user.Email != "" {
		req.Email = ""
	}
	if req.Phone != "" {
		if !w.VerifyCellphoneFormat(req.Phone) {
			return errors.New(w.Lang("手机号不正确"))
		}
		exist, err = w.GetUserInfoByPhone(req.Phone)
		if err == nil && exist.Id != user.Id {
			return errors.New(w.Lang("该手机号已被注册"))
		}
		user.Phone = req.Phone
	}
	if req.Email != "" {
		if !w.VerifyEmailFormat(req.Email) {
			return errors.New(w.Lang("邮箱不正确"))
		}
		exist, err = w.GetUserInfoByEmail(req.Email)
		if err == nil && exist.Id != user.Id {
			return errors.New(w.Lang("该邮箱已被注册"))
		}
		user.Email = req.Email
	}
	user.UserName = req.UserName
	user.RealName = req.RealName
	if user.GroupId == 0 {
		user.GroupId = w.PluginUser.DefaultGroupId
	}

	w.DB.Save(user)

	return nil
}

func (w *Website) CleanUserVip() {
	if w.DB == nil {
		return
	}
	var group model.UserGroup
	err := w.DB.Where("`status` = 1").Order("level asc").Take(&group).Error
	if err != nil {
		return
	}
	w.DB.Model(&model.User{}).Where("`status` = 1 and `group_id` != ? and `expire_time` < ?", group.Id, time.Now().Unix()).UpdateColumn("group_id", group.Id)
}

func (w *Website) GetUserDiscount(userId uint, user *model.User) int64 {
	if user == nil {
		user, _ = w.GetUserInfoById(userId)
	}
	if user != nil {
		if user.ParentId > 0 {
			parent, err := w.GetUserInfoById(user.ParentId)
			if err == nil {
				group, err := w.GetUserGroupInfo(parent.GroupId)
				if err == nil {
					if group.Setting.Discount > 0 {
						return group.Setting.Discount
					}
				}
			}
		}
	}

	return 0
}

func (w *Website) GetUserFields() []*config.CustomField {
	//这里有默认的设置
	fields := w.PluginUser.Fields

	return fields
}
