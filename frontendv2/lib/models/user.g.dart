// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

User _$UserFromJson(Map<String, dynamic> json) => User(
  id: (json['id'] as num?)?.toInt(),
  username: json['username'] as String,
  email: json['email'] as String,
  registeredAt: json['registered_at'] == null
      ? null
      : DateTime.parse(json['registered_at'] as String),
  job: json['job'] as String?,
  privilege: (json['privilege'] as num?)?.toInt(),
);

Map<String, dynamic> _$UserToJson(User instance) => <String, dynamic>{
  'id': instance.id,
  'username': instance.username,
  'email': instance.email,
  'registered_at': instance.registeredAt?.toIso8601String(),
  'job': instance.job,
  'privilege': instance.privilege,
};
