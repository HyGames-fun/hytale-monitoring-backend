import { IsEmail, IsNotEmpty, IsString, MinLength } from 'class-validator'
import { IsTrim } from '../../decorators/trim.decorator'
import { Transform } from 'class-transformer'

export class DiscordRegisterDto {
  name: string
  avatar: string
  discordId: string
}

export class RegisterDto {
  @IsTrim()
  @IsString()
  @IsNotEmpty()
  @MinLength(3)
  name: string

  @IsString()
  @IsNotEmpty()
  @IsEmail()
  @Transform((val) => (val.value as string).toLowerCase())
  email: string

  @IsString()
  @IsNotEmpty()
  @MinLength(6)
  password: string
}

export class LoginDto {
  @IsString()
  @IsNotEmpty()
  @IsEmail()
  @Transform((val) => (val.value as string).toLowerCase())
  email: string

  @IsString()
  @IsNotEmpty()
  password: string
}
