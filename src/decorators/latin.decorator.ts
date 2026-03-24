import { applyDecorators } from '@nestjs/common'
import { Validate } from 'class-validator'

export function IsLatin() {
  return applyDecorators(Validate(IsLatin))
}