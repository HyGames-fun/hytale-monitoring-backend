import { Region, Tag } from '../../../generated/prisma/enums'
import {
  ArrayNotEmpty,
  IsArray,
  IsEnum,
  IsNotEmpty,
  IsString,
} from 'class-validator'
import { IsDomain } from '../../decorators/domain.decorator'
import { IsIp } from '../../decorators/ip.decorator'
import { IsKebabCase } from '../../decorators/kebab-case.decorator'

export class ServerDto {
  id?: number
  ip?: string
  domain?: string
  nameId?: string
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
  domain: string

  @IsString()
  description: string

  @IsString()
  @IsNotEmpty()
  name: string

  @IsString()
  @IsNotEmpty()
  @IsKebabCase()
  nameId?: string

  @IsArray()
  @ArrayNotEmpty()
  @IsEnum(Tag, { each: true })
  tags: Tag[]

  @IsEnum(Region)
  region: Region
}
