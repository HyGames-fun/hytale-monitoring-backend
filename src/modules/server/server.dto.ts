import { Region, Tag } from '../../../generated/prisma/enums'
import {
  ArrayNotEmpty,
  IsArray,
  IsEnum,
  IsNotEmpty, IsOptional,
  IsString, Length,
} from 'class-validator'
import { IsDomain } from '../../decorators/domain.decorator'
import { IsIp } from '../../decorators/ip.decorator'
import { IsKebabCase } from '../../decorators/kebab-case.decorator'
import { Transform } from 'class-transformer'

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
  @IsOptional()
  @IsString()
  @IsIp()
  ip?: string

  @IsString()
  @IsDomain()
  domain: string

  @Transform(({ value }) => value?.trim())
  @IsString()
  @Length(10, 300)
  description: string

  @Transform(({ value }) => value?.trim())
  @IsString()
  @Length(2, 30)
  name: string

  @Transform(({ value }) => value?.trim())
  @IsString()
  @Length(2, 30)
  @IsKebabCase()
  nameId: string

  @IsArray()
  @ArrayNotEmpty()
  @IsEnum(Tag, { each: true })
  tags: Tag[]

  @IsEnum(Region)
  region: Region
}
