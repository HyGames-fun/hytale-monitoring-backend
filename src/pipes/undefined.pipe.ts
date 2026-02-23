import { BadRequestException, Injectable, PipeTransform } from '@nestjs/common'
import { ArgumentMetadata } from '@nestjs/common/interfaces/features/pipe-transform.interface'

@Injectable()
export class UndefinedPipe implements PipeTransform {
  transform(value: string, metadata?: ArgumentMetadata) {
    if (!value) throw new BadRequestException(`${metadata?.data} is undefined!`)
    return value
  }
}
