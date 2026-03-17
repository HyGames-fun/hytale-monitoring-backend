import { applyDecorators } from '@nestjs/common'
import { Validate } from 'class-validator'
import { TrimValidator } from '../validators/trim.validator'

export function IsTrim() {
  return applyDecorators(Validate(TrimValidator))
}
