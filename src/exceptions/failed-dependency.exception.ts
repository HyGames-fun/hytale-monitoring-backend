import { HttpException, HttpStatus } from '@nestjs/common'

export class FailedDependencyException extends HttpException {
  constructor(message = 'Failed dependency') {
    super(message, HttpStatus.FAILED_DEPENDENCY)
  }
}
