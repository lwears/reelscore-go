import { StarIcon } from '@heroicons/react/20/solid'
import { buildImgSrc } from '@web/lib/utils/helpers'
import clsx from 'clsx'
import Image from 'next/image'

export interface CardProps {
  posterPath: string | null
  title: string
  date: Date | null
  score?: number
  tmdbScore: number
  children?: React.ReactNode
}

// Add users score
export default function Card(data: CardProps) {
  const { posterPath, date, title, children, tmdbScore, score } = data
  return (
    <div className="shadow-lg hover:shadow-2xl group relative aspect-[2/3] w-full overflow-hidden rounded-xl border border-card-border text-sm font-extralight text-white hover:cursor-pointer transition-all duration-300 hover:scale-[1.02] hover:-translate-y-1">
      <div className="absolute left-0 top-0 z-0 size-full overflow-hidden rounded-xl bg-card group-hover:blur-sm transition-all duration-300 group-hover:scale-110">
        {posterPath ? (
          <Image
            src={buildImgSrc(posterPath)}
            alt={title || 'No Title'}
            width={300}
            height={320}
            className="size-full object-cover"
          />
        ) : (
          <p className="flex size-full items-center justify-center p-2 text-center text-xl font-semibold text-card-foreground">
            {title}
          </p>
        )}
      </div>
      <div className="flex h-full flex-col justify-between">
        <div
          className={clsx(
            'z-10 flex flex-1 flex-col justify-between p-3 opacity-0 group-hover:opacity-100 transition-all duration-300',
            posterPath
              ? 'bg-gradient-to-t from-black/95 via-transparent to-black/80'
              : 'bg-black/80'
          )}
        >
          <div className="flex flex-col gap-2">
            <div className="flex justify-between items-start gap-2">
              <p className="overflow-hidden text-ellipsis font-medium flex-1">
                {title}
                {date && ` (${date.getFullYear()})`}
              </p>
              <div className="flex flex-col gap-1 items-end">
                {score && Number(score) > 0 && (
                  <span className="flex gap-1.5 items-center bg-emerald-500/20 backdrop-blur-sm px-2 py-1 rounded-full border border-emerald-500/30">
                    <span className="text-[10px] font-medium text-emerald-300">
                      YOU
                    </span>
                    <p className="font-semibold text-emerald-100">
                      {Number(score)}
                    </p>
                    <StarIcon className="size-3.5 text-emerald-400" />
                  </span>
                )}
                {tmdbScore && Number(tmdbScore) > 0 && (
                  <span className="flex gap-1.5 items-center bg-amber-500/20 backdrop-blur-sm px-2 py-1 rounded-full border border-amber-500/30">
                    <span className="text-[10px] font-medium text-amber-300">
                      TMDB
                    </span>
                    <p className="font-semibold text-amber-100">
                      {Number(tmdbScore)}
                    </p>
                    <StarIcon className="size-3.5 text-amber-400" />
                  </span>
                )}
              </div>
            </div>
          </div>
          <div className="flex w-full">{children && children}</div>
        </div>
      </div>
    </div>
  )
}
