import { ServerDto } from '../server/server.dto'
import { Transform, Type } from 'class-transformer'
import {
  Allow,
  IsNotEmpty,
  IsNumber,
  Min,
  ValidateNested
} from 'class-validator'

export class UserDto {
  id?: number
  name?: string
  email?: string | null
  avatar?: string | null
  password?: string | null
  discordId?: string | null
  servers?: ServerDto[]
}

export class UpdateUserDto {
  name?: string
  email?: string | null
  avatar?: string | null
  password?: string | null
  discordId?: string | null
}

export class RegionDto {
  @IsNotEmpty()
  @IsNumber()
  @Min(0)
  left: number

  @IsNotEmpty()
  @IsNumber()
  @Min(0)
  top: number

  @IsNotEmpty()
  @IsNumber()
  @Min(1)
  width: number

  @IsNotEmpty()
  @IsNumber()
  @Min(1)
  height: number
}

export class UpdateAvatarDto {
  @Allow()
  @Transform(({ value }) => {
    try {
      return JSON.parse(value as string)
    } catch {
      throw new Error('Invalid JSON')
    }
  })
  @ValidateNested()
  @Type(() => RegionDto)
  region: RegionDto
}
