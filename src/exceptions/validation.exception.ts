import { HttpException, HttpStatus } from '@nestjs/common'
import { ValidationError } from 'class-validator'

export class ValidationException extends HttpException {
  constructor(errors: ValidationError[]) {
    super(ValidationException.buildResponse(errors), HttpStatus.BAD_REQUEST)
  }

  static buildResponse(errors: ValidationError[]): string | object {
    const formatErrors = errors.reduce((acc, error) => {
      acc[error.property] = Object.values(error.constraints as object)
      return acc
    }, {})
    return {
      status: 'validation_error',
      errors: formatErrors
    }
  }
}
