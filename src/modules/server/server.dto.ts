import { Region, Tag } from '../../../generated/prisma/enums'
import {
  IsArray,
  IsEnum,
  IsNotEmpty,
  IsOptional,
  IsString,
} from 'class-validator'
import { IsDomain } from '../../decorators/domain.decorator'
import { IsIp } from '../../decorators/ip.decorator'

export class ServerDto {
  id?: number
  ip?: string
  realIp?: string
  description?: string
  name?: string
  tags?: Tag[]
  region?: Region
}

export class CreateServerDto {
  @IsString()
  @IsNotEmpty()
  @IsDomain()
  ip: string

  @IsString()
  @IsNotEmpty()
  @IsIp()
  realIp: string

  @IsString()
  description: string

  @IsString()
  @IsNotEmpty()
  name: string

  @IsOptional()
  @IsArray()
  @IsEnum(Tag, { each: true })
  tags: Tag[]

  @IsEnum(Region)
  region: Region
}
