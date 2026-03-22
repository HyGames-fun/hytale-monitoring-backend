import { BadRequestException, Injectable, PipeTransform } from '@nestjs/common'
import { RegionDto } from '../user.dto'
import { plainToInstance } from 'class-transformer'
import { validateSync } from 'class-validator'
import { ValidationException } from '../../../exceptions/validation.exception'

@Injectable()
export class RegionPipe implements PipeTransform {
  transform(value: any) {
    if (!value || typeof value !== 'string')
      throw new BadRequestException('Invalid region data type')

    let parsed: unknown
    try {
      parsed = JSON.parse(value)
    } catch {
      throw new BadRequestException('Invalid JSON')
    }

    const region = plainToInstance(RegionDto, parsed)

    const errors = validateSync(region, { whitelist: true })

    if (errors.length) {
      throw new ValidationException(errors)
    }

    return region
  }
}
