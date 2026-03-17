import { IsEmail, IsNotEmpty, IsString, MinLength } from 'class-validator'
import { IsTrim } from '../../decorators/trim.decorator'

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
  email: string

  @IsString()
  @IsNotEmpty()
  password: string
}
