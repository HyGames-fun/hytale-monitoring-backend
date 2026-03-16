import { Region, Tag } from '../../../generated/prisma/enums'
import {
  ArrayNotEmpty,
  IsArray,
  IsEnum,
  IsIn,
  IsInt,
  IsNotEmpty,
  IsOptional,
  IsString,
  Length,
  Min,
  ValidateNested
} from 'class-validator'
import { IsDomain } from '../../decorators/domain.decorator'
import { IsIp } from '../../decorators/ip.decorator'
import { IsKebabCase } from '../../decorators/kebab-case.decorator'
import { Transform, Type } from 'class-transformer'

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
