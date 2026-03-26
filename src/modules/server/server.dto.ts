import { Region, Tag } from '../../../generated/prisma/enums'
import {
  ArrayNotEmpty,
  IsArray,
  IsEnum,
  IsIn,
  IsInt,
  IsOptional,
  IsString,
  IsUrl,
  Length,
  Min,
  ValidateNested
} from 'class-validator'
import { IsDomain } from '../../decorators/domain.decorator'
import { IsIp } from '../../decorators/ip.decorator'
import { IsKebabCase } from '../../decorators/kebab-case.decorator'
import { Type } from 'class-transformer'
import { IsTrim } from '../../decorators/trim.decorator'
import { AtLeastOne } from '../../decorators/at-least-one.decorator'
import { IsLatin } from '../../decorators/latin.decorator'

export class ServerDto {
  id?: number
  ip?: string | null
  domain?: string | null
  nameId?: string
  description?: string
  name?: string
  poster?: string | null
  tags?: Tag[]
  region?: Region
  userId?: number | null
  likesQuantity?: number
  createdAt?: Date
  updatedAt?: Date
}

class FiltersDto {
  @IsOptional()
  @IsArray()
  @IsEnum(Tag, { each: true })
  tags?: Tag[]

  @IsOptional()
  @IsEnum(Region)
  region?: Region
}

class OrderDto {
  @IsOptional()
  @IsIn(['asc', 'desc'])
  likes?: 'asc' | 'desc'

  @IsOptional()
  @IsIn(['asc', 'desc'])
  createdAt?: 'asc' | 'desc'
}

export class FindPageDto {
  @IsInt()
  @Min(0)
  page: number

  @IsInt()
  @Min(1)
  quantity: number

  @IsOptional()
  @ValidateNested()
  @Type(() => FiltersDto)
  filters?: FiltersDto

  @IsOptional()
  @ValidateNested()
  @Type(() => OrderDto)
  order?: OrderDto
}

export class CreateServerDto {
  @IsOptional()
  @IsString()
  @IsIp()
  ip?: string

  @IsOptional()
  @IsString()
  @IsDomain()
  domain?: string

  @AtLeastOne(['ip', 'domain'])
  private readonly _atLeastOne!: never

  @IsString()
  @Length(10, 300)
  @IsTrim()
  description: string

  @IsString()
  @Length(2, 30)
  @IsTrim()
  name: string

  @IsString()
  @Length(2, 30)
  @IsTrim()
  @IsKebabCase()
  nameId: string

  @IsOptional()
  @IsString()
  @IsUrl()
  poster?: string

  @IsArray()
  @ArrayNotEmpty()
  @IsEnum(Tag, { each: true })
  tags: Tag[]

  @IsEnum(Region)
  region: Region
}

export class UpdateServerDto {
  @IsOptional()
  @IsString()
  @IsIp()
  ip?: string

  @IsOptional()
  @IsString()
  @IsDomain()
  domain?: string

  @IsOptional()
  @IsString()
  @Length(10, 300)
  @IsTrim()
  description?: string

  @IsOptional()
  @IsString()
  @Length(2, 30)
  @IsTrim()
  name?: string

  @IsOptional()
  @IsString()
  @Length(2, 30)
  @IsTrim()
  @IsKebabCase()
  nameId?: string

  @IsOptional()
  @IsString()
  @IsUrl()
  poster?: string

  @IsOptional()
  @IsArray()
  @IsEnum(Tag, { each: true })
  tags?: Tag[]

  @IsOptional()
  @IsEnum(Region)
  region?: Region
}
