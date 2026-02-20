import { applyDecorators } from '@nestjs/common'
import { Validate } from 'class-validator'
import { IsDomainValidator } from '../validators/domain.validator'

export function IsDomain() {
  return applyDecorators(Validate(IsDomainValidator))
}
